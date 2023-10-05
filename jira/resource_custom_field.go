package jira

import (
	"log"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

type FieldResponse struct {
	Id          string `json:"id,omitempty" structs:"id,omitempty"`
	Name        string `json:"name,omitempty" structs:"name,omitempty"`
	Description string `json:"description,omitempty" structs:"description,omitempty"`
	SearcherKey string `json:"searcherKey,omitempty" structs:"searcherKey,omitempty"`
	Schema      struct {
		Type   string `json:"type,omitempty" structs:"type,omitempty"`
		Items  string `json:"items,omitempty" structs:"items,omitempty"`
		Custom string `json:"custom,omitempty" structs:"custom,omitempty"`
	} `json:"schema,omitempty" structs:"schema,omitempty"`
}

// FieldRequest The struct sent to the JIRA instance to create a new Field
type FieldRequest struct {
	Name        string `json:"name,omitempty" structs:"name,omitempty"`
	Description string `json:"description,omitempty" structs:"description,omitempty"`
	Type        string `json:"type,omitempty" structs:"type,omitempty"`
	SearcherKey string `json:"searcherKey,omitempty" structs:"searcherKey,omitempty"`
}

type GetFieldResponse struct {
	MaxResults int32           `json:"maxResults,omitempty" structs:"maxResults,omitempty"`
	StartAt    int32           `json:"startAt,omitempty" structs:"startAt,omitempty"`
	Total      int32           `json:"total,omitempty" structs:"total,omitempty"`
	IsLast     bool            `json:"isLast,omitempty" structs:"isLast,omitempty"`
	NextPage   string          `json:"nextPage,omitempty" structs:"nextPage,omitempty"`
	Self       string          `json:"self,omitempty" structs:"self,omitempty"`
	Values     []FieldResponse `json:"values,omitempty" structs:"values,omitempty"`
}

// resourceCustomField is used to define a JIRA custom field
func resourceCustomField() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomFieldCreate,
		Read:   resourceCustomFieldRead,
		Update: resourceCustomFieldUpdate,
		Delete: resourceCustomFieldDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"searcher_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getCustomFieldById(client *jira.Client, id string) (*FieldResponse, *GetFieldResponse, error) {
	returnedFields := new(GetFieldResponse)
	err := request(client, "GET", customFieldSearchEndpoint(id), nil, returnedFields)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Custom field lookup failed")
	}
	if returnedFields.Total != 1 {
		return nil, returnedFields, errors.Wrap(err, "Custom field not found")
	}

	return &returnedFields.Values[0], returnedFields, nil
}

// resourceCustomFieldRead reads custom field details using jira api
func resourceCustomFieldRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	field, _, err := getCustomFieldById(config.jiraClient, d.Id())
	if err != nil {
		return errors.Wrap(err, "Getting jira field failed")
	}

	log.Printf("Read custom field (id=%s)", field.Id)

	d.Set("name", field.Name)
	d.Set("description", field.Description)
	d.Set("type", field.Schema.Custom)
	d.Set("searcher_key", field.SearcherKey)

	return nil
}

// resourceCustomFieldCreate creates a new jira custom field using the jira api
func resourceCustomFieldCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	field := &FieldRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
		SearcherKey: d.Get("searcher_key").(string),
	}

	returnedField := new(jira.Field)

	err := request(config.jiraClient, "POST", fieldAPIEndpoint, field, returnedField)
	if err != nil {
		return errors.Wrap(err, "Request failed")
	}

	log.Printf("Created new Custom Field: %s", returnedField.ID)

	d.SetId(returnedField.ID)

	for err != nil {
		time.Sleep(200 * time.Millisecond)
		err = resourceCustomFieldRead(d, m)
	}

	return err
}

func resourceCustomFieldUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	field := &FieldRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		SearcherKey: d.Get("searcher_key").(string),
	}

	err := request(config.jiraClient, "PUT", customFieldEndpoint(d.Id()), field, nil)
	if err != nil {
		return errors.Wrap(err, "Update custom field request failed")
	}

	log.Printf("Updated Custom Field: %s", d.Id())

	return resourceCustomFieldRead(d, m)
}

func resourceCustomFieldDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	err := request(config.jiraClient, "DELETE", customFieldEndpoint(d.Id()), nil, nil)
	if err != nil {
		return errors.Wrap(err, "Deleting Jira Field Context Option failed")
	}

	return nil
}
