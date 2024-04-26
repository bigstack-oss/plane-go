package template

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
)

func Render(temlp string) (string, error) {
	template, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(temlp)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	var v interface{}
	err = template.Execute(&buf, v)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderByJob(job protocol.Job, temlp string) (string, error) {
	template, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(temlp)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, job)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
