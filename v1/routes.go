package v1

import (
	"github.com/faryon93/autosig/directory"
	"github.com/gorilla/mux"
	"net/http"
)

func Routes(r *mux.Router, directory *directory.Service, templateDir string) {
	r.Methods(http.MethodGet).Path("/{Template}").
		HandlerFunc(TemplateGen(directory, templateDir))
}
