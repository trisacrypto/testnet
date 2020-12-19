package trisads

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/trisacrypto/testnet/pkg/trisads/pb"
)

// VerifyContactEmail creates a verification token for each contact in the VASP contact
// list and sends them the verification email with instructions on how to verify their
// email address.
func (s *Server) VerifyContactEmail(vasp *pb.VASP) (err error) {
	// Create the verification tokens and save the VASP back to the database
	var contacts = []*pb.Contact{
		vasp.Contacts.Technical, vasp.Contacts.Administrative,
		vasp.Contacts.Billing, vasp.Contacts.Legal,
	}

	for _, contact := range contacts {
		if contact != nil && contact.Email != "" {
			contact.Token = CreateToken(48)
			contact.Verified = false
		}
	}

	if err = s.db.Update(vasp); err != nil {
		log.Error().Msg("could not update vasp")
		return err
	}

	for _, contact := range contacts {
		if contact == nil || contact.Email == "" {
			continue
		}

		ctx := verifyContactContext{
			Name:  contact.Name,
			Token: contact.Token,
			VID:   vasp.Id,
		}

		var text, html string
		if text, err = execTemplateString(verifyContactPlainText, ctx); err != nil {
			return err
		}
		if html, err = execTemplateString(verifyContactHTML, ctx); err != nil {
			return err
		}

		if err = s.sendEmail(contact.Name, contact.Email, verifyContactSubject, text, html); err != nil {
			return err
		}

	}

	return nil
}

// ReviewRequestEmail is a shortcut for iComply verification in which we simply send
// an email to the TRISA admins and have them manually verify registrations.
func (s *Server) ReviewRequestEmail(vasp *pb.VASP) (err error) {
	// Create verification token for admin and update database
	// TODO: replace with actual authentication
	vasp.AdminVerificationToken = CreateToken(48)
	if err = s.db.Update(vasp); err != nil {
		return fmt.Errorf("could not save admin verification token: %s", err)
	}

	var data []byte
	if data, err = json.MarshalIndent(vasp, "", "  "); err != nil {
		return err
	}

	ctx := reviewRequestContext{
		VID:     vasp.Id,
		Request: string(data),
		Token:   vasp.AdminVerificationToken,
	}

	var text, html string
	if text, err = execTemplateString(reviewRequestPlainText, ctx); err != nil {
		return err
	}
	if html, err = execTemplateString(reviewRequestHTML, ctx); err != nil {
		return err
	}

	if err = s.sendEmail("TRISA Admins", s.conf.AdminEmail, reviewRequestSubject, text, html); err != nil {
		return err
	}

	return nil
}

func (s *Server) sendEmail(recipient, emailaddr, subject, text, html string) (err error) {
	message := mail.NewSingleEmail(
		mail.NewEmail("TRISA Directory Service", s.conf.ServiceEmail),
		subject,
		mail.NewEmail(recipient, emailaddr),
		text, html,
	)

	var rep *rest.Response
	if rep, err = s.email.Send(message); err != nil {
		return err
	}

	if rep.StatusCode < 200 || rep.StatusCode >= 300 {
		return errors.New(rep.Body)
	}

	return nil
}

var chars = []rune("ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz1234567890")

// CreateToken creates a variable length random token that can be used for passwords or API keys.
func CreateToken(length int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[random.Intn(len(chars))])
	}
	return b.String()
}

type verifyContactContext struct {
	Name  string
	Token string
	VID   string
}

var verifyContactSubject = "Verify Email Address"

// VerifyContact Plain Text Content Template
var verifyContactPlainText = template.Must(template.New("verifyContactPlainText").Parse(`
Hello {{ .Name }},

Thank you for submitting a TRISA TestNet VASP registration request. To begin the
verification process, please submit the following email verification token using the
VerifyEmail RPC in the TRISA directory service protocol:

ID: {{ .VID }}
Token: {{ .Token }}

This can be done with the trisads CLI utility or using the protocol buffers library in
the programming language of your choice.

Note that we're working to create a URL endpoint for the vaspdirectory.net site to
simplify the verification process. We're sorry about the inconvenience of this method at
the early stage of the TRISA Test Net.

Best Regards,
The TRISA Directory Service`))

// VerifyContact HTML Content Template
var verifyContactHTML = template.Must(template.New("verifyContactHTML").Parse(`
<p>Hello {{ .Name }},</p>

<p>Thank you for submitting a TRISA TestNet VASP registration request. To begin the
verification process, please submit the following email verification token using the
VerifyEmail RPC in the TRISA directory service protocol:</p>

<ul>
	<li>ID: <strong>{{ .VID }}</strong></li>
	<li>Token: <strong>{{ .Token }}</strong></li>
</ul>

<p>This can be done with the trisads CLI utility or using the protocol buffers library in
the programming language of your choice.</p>

<p>Note that we're working to create a URL endpoint for the
<a href="https://vaspdirectory.net/">vaspdirectory.net</a> site to simplify the
verification process. We're sorry about the inconvenience of this method at the early
stage of the TRISA Test Net.</p>

<p>Best Regards,<br />
The TRISA Directory Service</p>`))

type reviewRequestContext struct {
	VID     string
	Token   string
	Request string
}

var reviewRequestSubject = "Please Review TRISA TestNET VASP Registration Request"

// VerifyContact Plain Text Content Template
var reviewRequestPlainText = template.Must(template.New("reviewRequestPlainText").Parse(`
Hello TRISA Admin,

We have received a new registration request from a VASP that needs to be reviewed. The
requestor has verified their email address and received a PKCS12 password to decrypt a
certificate that will be generated if you approve this request. The request JSON is:

{{ .Request }}

To verify or reject the registration request, use the following metadata with the
trisads verify command:

ID: {{ .VID }}
Token: {{ .Token }}

Note that we're working to create a URL endpoint for the vaspdirectory.net site to
simplify the verification process. We're sorry about the inconvenience of this method at
the early stage of the TRISA Test Net.

Best Regards,
The TRISA Directory Service`))

// VerifyContact HTML Content Template
var reviewRequestHTML = template.Must(template.New("reviewRequestHTML").Parse(`
<p>Hello TRISA Admin,</p>

<p>We have received a new registration request from a VASP that needs to be reviewed.
The requestor has verified their email address and received a PKCS12 password to decrypt
a certificate that will be generated if you approve this request. The request JSON is:</p>

<pre>{{ .Request }}</pre>

<p>To verify or reject the registration request, use the following metadata with the
<code>trisads verify</code> command:</p>

<ul>
	<li>ID: <strong>{{ .VID }}</strong></li>
	<li>Token: <strong>{{ .Token }}</strong></li>
</ul>

<p>Note that we're working to create a URL endpoint for the
<a href="https://vaspdirectory.net/">vaspdirectory.net</a> site to simplify the
verification process. We're sorry about the inconvenience of this method at the early
stage of the TRISA TestNet.</p>

<p>Best Regards,<br />
The TRISA Directory Service</p>`))

func execTemplateString(t *template.Template, ctx interface{}) (_ string, err error) {
	buf := new(strings.Builder)
	if err = t.Execute(buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}
