package protocol

import (
	json "github.com/json-iterator/go"
)

type Job struct {
	ID         string `json:"id"`
	Version    int    `json:"version"`
	*Applicant `json:"applicant"`
	*Desired   `json:"desired"`
	*Result    `json:"result"`
}

type Applicant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Project  string `json:"project"`
	Email    string `json:"email,omitempty"`
	Country  string `json:"country,omitempty"`
	Company  string `json:"company,omitempty"`
	Industry string `json:"industry,omitempty"`
}

type Desired struct {
	Operation string `json:"operation"`
	*Resource `json:"resource"`
}

type Resource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type Networks struct {
}

type Result struct {
	Status string `json:"status"`
	Desc   string `json:"desc"`
	Data   []byte `json:"data"`
}

func (j *Job) Reset() {
	j.Applicant = nil
	j.Desired = nil
	j.Result = nil
}

func (j *Job) String() string {
	b, err := json.Marshal(j)
	if err != nil {
		return ""
	}

	return string(b)
}

func (j *Job) Bytes() []byte {
	b, err := json.Marshal(j)
	if err != nil {
		return nil
	}

	return b
}
