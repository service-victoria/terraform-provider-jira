package jira

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

type FieldOption struct {
	Id       string `json:"id,omitempty" structs:"id,omitempty"`
	Value    string `json:"value,omitempty" structs:"value,omitempty"`
	Disabled bool   `json:"disabled,omitempty" structs:"disabled,omitempty"`
	OptionId string `json:"optionId,omitempty" structs:"option_id,omitempty"`
}

// resourceCustomField is used to define a JIRA custom field's options
func resourceCustomFieldOption() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomFieldOptionCreate,
		Read:   resourceCustomFieldOptionRead,
		Update: resourceCustomFieldOptionUpdate,
		Delete: resourceCustomFieldOptionDelete,

		Schema: map[string]*schema.Schema{
			"context_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"field_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"disabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"option_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceCustomFieldOptionCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	createOptionResponse := new(FieldOptions)
	createOptionRequest := &FieldOptions{
		Options: []FieldOption{
			FieldOption{
				Value:    d.Get("value").(string),
				Disabled: d.Get("disabled").(bool),
				OptionId: d.Get("option_id").(string),
			},
		},
	}
	out, _ := json.Marshal(createOptionRequest)
	log.Printf("Creating new option: %s", out)
	err := request(config.jiraClient, "POST", customFieldContextOptionsEndpoint(fieldId, contextId), createOptionRequest, createOptionResponse)
	if err != nil {
		return errors.Wrap(err, "Creating Jira Field Option failed")
	}

	d.SetId(createOptionResponse.Options[0].Id)

	return resourceCustomFieldOptionRead(d, m)
}

func resourceCustomFieldOptionRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	optionResponse := new(GetFieldOptionsResponse)
	err := request(config.jiraClient, "GET", customFieldOptionEndpoint(fieldId, contextId, d.Id()), nil, optionResponse)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Option failed")
	}

	option := optionResponse.Values[0]

	d.Set("value", option.Value)
	d.Set("disabled", option.Disabled)
	d.Set("option_id", option.OptionId)

	return nil
}

func resourceCustomFieldOptionUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	updateOptionsResponse := new(FieldOptions)

	updateOptionsRequest := &FieldOptions{
		Options: []FieldOption{
			FieldOption{
				Id:       d.Id(),
				Value:    d.Get("value").(string),
				Disabled: d.Get("disabled").(bool),
				OptionId: d.Get("option_id").(string),
			},
		},
	}
	out, _ := json.Marshal(updateOptionsRequest)
	log.Printf("Updating existing option: %s", out)
	err := request(config.jiraClient, "PUT", customFieldContextOptionsEndpoint(fieldId, contextId), updateOptionsRequest, updateOptionsResponse)
	if err != nil {
		return errors.Wrap(err, "Updating Jira Field Option failed")
	}

	return resourceCustomFieldOptionsRead(d, m)
}

func resourceCustomFieldOptionDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	err := request(config.jiraClient, "DELETE", customFieldContextOptionsDeleteEndpoint(fieldId, contextId, d.Id()), nil, nil)
	if err != nil {
		return errors.Wrap(err, "Deleting Jira Field Context Option failed")
	}

	return nil
}
