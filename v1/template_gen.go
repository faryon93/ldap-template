package v1

import (
	"github.com/faryon93/ldap-template/directory"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	reTemplateName = regexp.MustCompile(`^[A-Za-z0-9-]*$`)
)

func TemplateGen(dirService *directory.Service, templateDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		templateName := mux.Vars(r)["Template"]
		log := logrus.
			WithField("handler", "TemplateGen").
			WithField("template", templateName)

		if !reTemplateName.MatchString(templateName) {
			http.Error(w, "malformed template name: only numbers, chars and dashes are allowed", http.StatusBadRequest)
			log.Warnln("rejecting request: malformed template Name")
			return
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "query parameter 'username' missing", http.StatusBadRequest)
			log.Warnln("rejecting request: query paramter 'username' is missing")
			return
		}

		person, err := dirService.GetPerson(username)
		if err == directory.ErrPersonNotFound {
			http.Error(w, "person not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Errorln("person lookup failed:", err.Error())
			return
		}

		templateType := "text"
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			templateType = "html"
		}

		templatePath := filepath.Join(templateDir, templateName+"."+templateType+".tmpl")
		fh, err := os.Open(templatePath)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Errorln("failed to open template:", err.Error())
			return
		}
		buf, err := io.ReadAll(fh)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Errorln("failed to read template:", err.Error())
			return
		}

		tpl, err := template.New("body").Parse(string(buf))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Errorln("failed to parse template:", err.Error())
			return
		}

		// render template
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if templateType == "text" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		err = tpl.Execute(w, person)
		if err != nil {
			log.Println(err)
			http.Error(w, "failed to parse template", http.StatusInternalServerError)
		}
	}
}
