package service

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"necutya/faker/internal/domain/domain"
)

const (
	emailConfirmTemplate  = "confirm-email.tmpl"
	passwordResetTemplate = "password-reset.tmpl"
	orderSuccessTemplate  = "order-success.tmpl"
	planUpdateTemplate    = "plan-update.tmpl"
)

var (
	onceNotification sync.Once
	templates        *template.Template
)

func parseNotificationTemplates() (*template.Template, error) {
	var err error

	onceNotification.Do(func() {
		var files []os.FileInfo

		files, err = ioutil.ReadDir("internal/templates")
		if err != nil {
			return
		}

		templateFiles := make([]string, 0, len(files))

		for _, file := range files {
			filename := file.Name()

			if strings.HasSuffix(filename, ".tmpl") {
				templateFiles = append(templateFiles, "internal/templates/"+filename)
			}
		}

		templates, err = template.ParseFiles(templateFiles...)
	})

	return templates, err
}

func getEmailConfirmationTemplate(name, code string, codeTTL int64) (string, error) {
	tmpls, err := parseNotificationTemplates()
	if err != nil {
		return "", err
	}

	emailConfirmationStruct := struct {
		Name    string
		Code    string
		CodeTTL int
	}{
		Name:    name,
		Code:    code,
		CodeTTL: int((time.Duration(codeTTL) * time.Second).Minutes()),
	}

	t := tmpls.Lookup(emailConfirmTemplate)

	sw := bytes.NewBufferString("")

	if err := t.Execute(sw, emailConfirmationStruct); err != nil {
		return "", err
	}

	return sw.String(), nil
}

func getPasswordResetTemplate(name, code string, codeTTL int64) (string, error) {
	tmpls, err := parseNotificationTemplates()
	if err != nil {
		return "", err
	}

	passwordResetInfo := struct {
		Name    string
		Code    string
		CodeTTL int
	}{
		Name:    name,
		Code:    code,
		CodeTTL: int((time.Duration(codeTTL) * time.Second).Minutes()),
	}

	t := tmpls.Lookup(passwordResetTemplate)

	sw := bytes.NewBufferString("")

	if err := t.Execute(sw, passwordResetInfo); err != nil {
		return "", err
	}

	return sw.String(), nil
}

func getOrderSuccessTemplate(name, planName string, planEndDate *time.Time) (string, error) {
	tmpls, err := parseNotificationTemplates()
	if err != nil {
		return "", err
	}

	orderSuccessInfo := struct {
		Name        string
		PlanName    string
		PlanEndDate string
	}{
		Name:        name,
		PlanName:    planName,
		PlanEndDate: planEndDate.Format(domain.DateLayout),
	}

	if planEndDate == nil {
		orderSuccessInfo.PlanEndDate = ""
	} else {
		orderSuccessInfo.PlanEndDate = planEndDate.Format(domain.DateLayout)
	}

	t := tmpls.Lookup(orderSuccessTemplate)

	sw := bytes.NewBufferString("")

	if err := t.Execute(sw, orderSuccessInfo); err != nil {
		return "", err
	}

	return sw.String(), nil
}

func getPlanUpdateTemplate(name string) (string, error) {
	tmpls, err := parseNotificationTemplates()
	if err != nil {
		return "", err
	}

	planUpdateInfo := struct {
		Name string
	}{
		Name: name,
	}

	t := tmpls.Lookup(planUpdateTemplate)

	sw := bytes.NewBufferString("")

	if err := t.Execute(sw, planUpdateInfo); err != nil {
		return "", err
	}

	return sw.String(), nil
}
