package jira

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// FieldRequest The struct sent to the JIRA instance to create a new Field
type FieldContextRequest struct {
	Name         string   `json:"name,omitempty" structs:"name,omitempty"`
	Description  string   `json:"description,omitempty" structs:"description,omitempty"`
	IssueTypeIds []string `json:"issueTypeIds,omitempty" structs:"issueTypeIds,omitempty"`
	ProjectIds   []string `json:"projectIds,omitempty" structs:"projectIds,omitempty"`
}

type FieldContextResponse struct {
	Id           string   `json:"id" structs:"id"`
	Name         string   `json:"name" structs:"name"`
	Description  string   `json:"description,omitempty" structs:"description,omitempty"`
	IssueTypeIds []string `json:"issueTypeIds,omitempty" structs:"issueTypeIds,omitempty"`
	ProjectIds   []string `json:"projectIds,omitempty" structs:"projectIds,omitempty"`
}

// resourceCustomField is used to define a JIRA custom field context
func resourceCustomFieldContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomFieldContextCreate,
		Read:   resourceCustomFieldContextRead,
		Update: resourceCustomFieldContextUpdate,
		Delete: resourceCustomFieldContextDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"field_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"issue_type_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
			},
			"project_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
			},
			"is_any_issue_type": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: false,
				Required: false,
				Computed: true,
			},
			"is_global_context": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: false,
				Required: false,
				Computed: true,
			},
		},
	}
}

func resourceCustomFieldContextCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	context := &FieldContextRequest{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		IssueTypeIds: d.Get("issue_type_ids").([]string),
		ProjectIds:   d.Get("project_ids").([]string),
	}

	fieldId := d.Get("field_id").(string)
	returnedContext := new(FieldContextResponse)
	err := request(config.jiraClient, "POST", customFieldContextEndpoint(fieldId), context, returnedContext)
	if err != nil {
		return errors.Wrap(err, "Creating Jira Field Context failed")
	}

	d.SetId(returnedContext.Id)

	return resourceCustomFieldContextRead(d, m)
}

func resourceCustomFieldContextRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	returnedContexts := new(FieldContextsResponse)
	err := request(config.jiraClient, "GET", customFieldContextEndpoint(fieldId), nil, returnedContexts)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Contexts failed")
	}

	id := d.Id()

	if id == "" {
		return errors.New("Context ID not set")
	} else {
		for _, context := range returnedContexts.Values {
			if context.Id == id {
				log.Printf("Context found(id=%s), setting values", context.Id)
				setContextFields(d, context)
			}
		}
	}

	return nil
}

func resourceCustomFieldContextUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	context := &FieldContextRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	fieldId := d.Get("field_id").(string)
	err := request(config.jiraClient, "PUT", customFieldContextUpdateEndpoint(fieldId, d.Id()), context, nil)
	if err != nil {
		return errors.Wrap(err, "Updating Jira Field Context failed")
	}

	return resourceCustomFieldContextRead(d, m)
}

func resourceCustomFieldContextDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	err := request(config.jiraClient, "DELETE", customFieldContextUpdateEndpoint(fieldId, d.Id()), nil, nil)
	if err != nil {
		return errors.Wrap(err, "Updating Jira Field Context failed")
	}

	return nil
}
