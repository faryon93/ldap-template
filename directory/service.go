package directory

import (
	"errors"
	"github.com/go-ldap/ldap"
)

var (
	ErrPersonNotFound = errors.New("person not found")
)

const (
	TimeFormatLdap = "20060102150405Z"
)

type Service struct {
	PersonSearchBaseDn string
	ldapUrl            string
	ldapUser           string
	ldapPassword       string
}

func NewService(ldapUrl, ldapUser, ldapPassword string) *Service {
	return &Service{
		ldapUrl:      ldapUrl,
		ldapUser:     ldapUser,
		ldapPassword: ldapPassword,
	}
}

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
