package jira

import (
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

	id := d.Id()

	if id != "" {
		for _, context := range returnedContexts.Values {
			if context.Id == id {
				log.Printf("Context found(id=%s), setting values", context.Id)
				setContextFields(d, context)
			}
		}
	} else if len(returnedContexts.Values) != 1 {
		return errors.New("Unable to determine context from field ID. Are there multiple contexts defined?")
	} else {
		context := returnedContexts.Values[0]
		log.Printf("Context found(id=%s), setting values", context.Id)
		setContextFields(d, context)
	}

	return nil
}

func setContextFields(d *schema.ResourceData, context FieldContext) {
	d.SetId(context.Id)
	d.Set("name", context.Name)
	d.Set("description", context.Description)
	d.Set("is_any_issue_type", context.IsAnyIssueType)
	d.Set("is_global_context", context.IsGlobalContext)
}
