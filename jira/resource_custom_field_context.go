package jira

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

const ID_FORMAT_ERROR_MESSAGE = "ID is incorrectly formatted. Expected format is `field_id:context_id`"

// FieldRequest The struct sent to the JIRA instance to create a new Field
type FieldContextRequest struct {
	Name         string   `json:"name,omitempty" structs:"name,omitempty"`
	Description  string   `json:"description" structs:"description"`
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
			"context_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: false,
				Computed: true,
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

	d.Set("context_id", returnedContext.Id)
	d.SetId(fmt.Sprintf("%s:%s", fieldId, returnedContext.Id))

	return resourceCustomFieldContextRead(d, m)
}

func resourceCustomFieldContextRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return errors.New(ID_FORMAT_ERROR_MESSAGE)
	}
	fieldId, contextId := ids[0], ids[1]

	if fieldId == "" || contextId == "" {
		return errors.New(ID_FORMAT_ERROR_MESSAGE)
	}

	returnedContexts := new(FieldContextsResponse)
	err := request(config.jiraClient, "GET", customFieldContextEndpoint(fieldId), nil, returnedContexts)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Contexts failed")
	}

	for _, context := range returnedContexts.Values {
		if context.Id == contextId {
			log.Printf("Context found(id=%s), setting values", context.Id)
			setContextFields(d, context, fieldId)
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

	ids := strings.Split(d.Id(), ":")
	fieldId, contextId := ids[0], ids[1]
	err := request(config.jiraClient, "PUT", customFieldContextUpdateEndpoint(fieldId, contextId), context, nil)
	if err != nil {
		return errors.Wrap(err, "Updating Jira Field Context failed")
	}

	return resourceCustomFieldContextRead(d, m)
}

func resourceCustomFieldContextDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	ids := strings.Split(d.Id(), ":")
	fieldId, contextId := ids[0], ids[1]
	err := request(config.jiraClient, "DELETE", customFieldContextUpdateEndpoint(fieldId, contextId), nil, nil)
	if err != nil {
		return errors.Wrap(err, "Updating Jira Field Context failed")
	}

	return nil
}
