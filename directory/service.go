package directory

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
	"errors"
	"fmt"
	"time"

	"github.com/go-ldap/ldap"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------------------
//  global variables
// ---------------------------------------------------------------------------------------

var (
	ErrPersonNotFound = errors.New("person not found")
)

// ---------------------------------------------------------------------------------------
//  constants
// ---------------------------------------------------------------------------------------

const (
	TimeFormatLdap = "20060102150405Z"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Service struct {
	PersonSearchBaseDn string
	ldapUrl            string
	ldapUser           string
	ldapPassword       string
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

func NewService(ldapUrl, ldapUser, ldapPassword string) *Service {
	return &Service{
		ldapUrl:      ldapUrl,
		ldapUser:     ldapUser,
		ldapPassword: ldapPassword,
	}
}

// ---------------------------------------------------------------------------------------
//  private methods
// ---------------------------------------------------------------------------------------

// newConn creates a new ldap connection and connects to the server.
func (s *Service) newConn() (*ldap.Conn, error) {
	l, err := ldap.DialURL(s.ldapUrl)
	if err != nil {
		return nil, err
	}

	err = l.Bind(s.ldapUser, s.ldapPassword)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// ---------------------------------------------------------------------------------------
//  public methods
// ---------------------------------------------------------------------------------------

// GetPerson searches for the given person in the directory.
func (s *Service) GetPerson(username string) (*Person, error) {
	log := logrus.
		WithField("action", "lookup-person").
		WithField("username", username)

	conn, err := s.newConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// serach query
	searchRequest := ldap.NewSearchRequest(
		s.PersonSearchBaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(samAccountName=%s))", ldap.EscapeFilter(username)),
		[]string{"dn", "displayName", "mail", "telephoneNumber", "description", "otherTelephone", "whenChanged"},
		nil,
	)

	searchResult, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	numPersons := len(searchResult.Entries)
	if numPersons > 1 {
		log.Warnf("directory search returned %d results, only one expected: aborting", numPersons)
	} else if numPersons < 1 {
		log.Warnln("no person found for the given username")
		return nil, ErrPersonNotFound
	}

	// ldap object to Person object
	entry := searchResult.Entries[0]

	timeChanged, err := time.Parse(TimeFormatLdap, entry.GetAttributeValue("whenChanged"))
	if err != nil {
		return nil, err
	}

	person := Person{
		DisplayName:       entry.GetAttributeValue("displayName"),
		Description:       entry.GetAttributeValue("description"),
		Mail:              entry.GetAttributeValue("mail"),
		TelephoneNumber:   entry.GetAttributeValue("telephoneNumber"),
		MobilephoneNumber: entry.GetAttributeValue("otherTelephone"),
		TimeChanged:       timeChanged,
	}

	if person.DisplayName == "" {
		log.Warnln("person has no display name: aborting")
		return nil, ErrPersonNotFound
	}

	return &person, nil
}
