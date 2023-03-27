package directory

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/sirupsen/logrus"
	"time"
)

type Person struct {
	DisplayName       string
	Description       string
	Mail              string
	TelephoneNumber   string
	MobilephoneNumber string

	TimeChanged time.Time
}

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
