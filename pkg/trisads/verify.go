package trisads

import (
	"encoding/json"
	"errors"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/trisacrypto/testnet/pkg/trisads/pb"
)

// SendVerificationEmail is a shortcut for iComply verification in which we simply send
// an email to the TRISA admins and have them manually verify registrations.
func (s *Server) SendVerificationEmail(vasp pb.VASP) (err error) {
	from := mail.NewEmail("TRISA Directory Service", s.conf.ServiceEmail)
	subject := "TRISA Test Net Verification Request"
	to := mail.NewEmail("TRISA Admins", s.conf.AdminEmail)

	var data []byte
	if data, err = json.MarshalIndent(vasp, "", "  "); err != nil {
		return err
	}
	plainTextContent := string(data)
	htmlContent := "<pre>" + plainTextContent + "</pre>"

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	var rep *rest.Response
	if rep, err = s.email.Send(message); err != nil {
		return err
	}

	if rep.StatusCode < 200 || rep.StatusCode >= 300 {
		return errors.New(rep.Body)
	}

	return nil
}
