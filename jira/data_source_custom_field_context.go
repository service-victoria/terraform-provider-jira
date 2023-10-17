package jira

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

type FieldContext struct {
	Id              string `json:"id" structs:"id"`
	Name            string `json:"name" structs:"name"`
	Description     string `json:"description" structs:"description"`
	IsAnyIssueType  bool   `json:"isAnyIssueType" structs:"isAnyIssueType"`
	IsGlobalContext bool   `json:"isGlobalContext" structs:"isGlobalContext"`
}

type FieldContextsResponse struct {
	MaxResults int32          `json:"maxResults" structs:"maxResults"`
	StartAt    int32          `json:"startAt" structs:"startAt"`
	Total      int32          `json:"total" structs:"total"`
	IsLast     bool           `json:"isLast" structs:"isLast"`
	NextPage   string         `json:"nextPage" structs:"nextPage"`
	Self       string         `json:"self" structs:"self"`
	Values     []FieldContext `json:"values" structs:"values"`
}

// resourceCustomField is used to define a JIRA custom field context
func dataSourceCustomFieldContext() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCustomFieldContextRead,

		Schema: map[string]*schema.Schema{
			"field_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"context_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_any_issue_type": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_global_context": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomFieldContextRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	returnedContexts := new(FieldContextsResponse)
	err := request(config.jiraClient, "GET", customFieldContextEndpoint(fieldId), nil, returnedContexts)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Contexts failed")
	}

	contextId := d.Get("context_id").(string)

	var foundContext FieldContext
	if contextId != "" {
		for _, context := range returnedContexts.Values {
			if context.Id == contextId {
				foundContext = context
			}
		}
	} else if len(returnedContexts.Values) == 1 {
		foundContext = returnedContexts.Values[0]
	}

	if foundContext == (FieldContext{}) {
		return errors.New("Unable to determine context from field ID. Are there multiple contexts defined?")
	}

	log.Printf("Context found(id=%s), setting values", foundContext.Id)
	d.SetId(fmt.Sprintf("%s:%s", fieldId, foundContext.Id))
	setContextFields(d, foundContext, fieldId)

	return nil
}

func setContextFields(d *schema.ResourceData, context FieldContext, fieldId string) {
	d.Set("name", context.Name)
	d.Set("field_id", fieldId)
	d.Set("context_id", context.Id)
	d.Set("description", context.Description)
	d.Set("is_any_issue_type", context.IsAnyIssueType)
	d.Set("is_global_context", context.IsGlobalContext)
}
