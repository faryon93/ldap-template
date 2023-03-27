package main

// ldap-template
// Copyright (C) 2023 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"net/http"
	"os"
	"syscall"

	"github.com/faryon93/util"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"

	"github.com/faryon93/ldap-template/directory"
	"github.com/faryon93/ldap-template/v1"
)

// ---------------------------------------------------------------------------------------
//  global variables
// ---------------------------------------------------------------------------------------

var (
	LdapUrl      string
	LdapUsername string
	LdapPassword string
	LdapSearchDn string
	HttpListen   string
	TemplateDir  string
)

// ---------------------------------------------------------------------------------------
//  application entry
// ---------------------------------------------------------------------------------------

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

	if LdapUrl == "" {
		logrus.Errorln("--ldap-url should not be omitted")
		os.Exit(1)
	}

	if LdapSearchDn == "" {
		logrus.Errorln("--ldap-search-dn should not be omitted")
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
