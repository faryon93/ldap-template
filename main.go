package main

import (
	"net/http"
	"os"
	"syscall"

	"github.com/faryon93/autosig/directory"
	v1 "github.com/faryon93/autosig/v1"

	"github.com/faryon93/util"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
)

var (
	LdapUrl      string
	LdapUsername string
	LdapPassword string
	LdapSearchDn string
	HttpListen   string
	TemplateDir  string
)

func main() {
	flags := flag.NewFlagSetWithEnvPrefix("main", "AUTOSIG", flag.ExitOnError)
	flags.StringVar(&LdapUrl, "ldap-url", "", "URL of the ldap server (e.g.: ldap://localhost:389)")
	flags.StringVar(&LdapUsername, "ldap-user", "", "Username for LDAP bind (e.g.: cn=ldapsearch,cn=Users,dc=example,dc=com")
	flags.StringVar(&LdapPassword, "ldap-password", "", "Password for LDAP user")
	flags.StringVar(&LdapSearchDn, "ldap-search-dn", "", "person search DN (e.b.: cn=%s,cn=Users,dc=example,dc=com")
	flags.StringVar(&HttpListen, "http-listen", ":8000", "")
	flags.StringVar(&TemplateDir, "template-dir", "templates/", "")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		logrus.Errorln("failed to parse command line:", err.Error())
		os.Exit(1)
	}

	logrus.Infoln("ldap directory server:", LdapUrl)
	dirService := directory.NewService(LdapUrl, LdapUsername, LdapPassword)
	dirService.PersonSearchBaseDn = LdapSearchDn

	router := mux.NewRouter()
	v1.Routes(router.PathPrefix("/v1").Subrouter(), dirService, TemplateDir)

	go func() {
		logrus.Infoln("http listening on", HttpListen)
		srv := &http.Server{Addr: HttpListen, Handler: router}
		err := srv.ListenAndServe()
		if err != nil {
			logrus.Errorln("http listen failed:", err.Error())
			os.Exit(1)
		}
	}()

	util.WaitSignal(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logrus.Infoln("shutting down application")
}
