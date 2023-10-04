package jira

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// FieldOptions The struct sent to the JIRA instance to create a new Field
type FieldOptions struct {
	Options []FieldOption `json:"options" structs:"options"`
}

type GetFieldOptionsResponse struct {
	Values []FieldOption `json:"values" structs:"values"`
}

type ReorderFieldOptionsRequest struct {
	Position             string   `json:"position" structs:"position"`
	CustomFieldOptionIds []string `json:"customFieldOptionIds" structs:"customFieldOptionIds"`
}

// resourceCustomFieldOptions is used to define a JIRA custom field's options
func resourceCustomFieldOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomFieldOptionsCreate,
		Read:   resourceCustomFieldOptionsRead,
		Update: resourceCustomFieldOptionsUpdate,
		Delete: resourceCustomFieldOptionsDelete,

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
			"options_ids": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
		},
	}
}

func resourceCustomFieldOptionsCreate(d *schema.ResourceData, m interface{}) error {
	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)
	d.SetId(fmt.Sprintf("%s-%s", fieldId, contextId))

	return resourceCustomFieldOptionsUpdate(d, m)
}

func resourceCustomFieldOptionsRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	currentOptions := new(GetFieldOptionsResponse)
	err := request(config.jiraClient, "GET", customFieldContextOptionsEndpoint(fieldId, contextId), nil, currentOptions)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Context Options failed")
	}

	optionIds := []string{}
	for _, opt := range currentOptions.Values {
		optionIds = append(optionIds, opt.Id)
	}

	d.Set("options_ids", optionIds)

	return nil
}

func resourceCustomFieldOptionsUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	resourceOptionIds := d.Get("options_ids").([]interface{})
	optionIds := make([]string, 0, len(resourceOptionIds))
	for _, id := range resourceOptionIds {
		optionIds = append(optionIds, id.(string))
	}

	currentOptions := new(GetFieldOptionsResponse)
	err := request(config.jiraClient, "GET", customFieldContextOptionsEndpoint(fieldId, contextId), nil, currentOptions)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Context Options failed")
	}

	for _, option := range currentOptions.Values {
		exists := optionsIdsContains(optionIds, option)
		if !exists {
			out, _ := json.Marshal(option)
			log.Printf("Deleting option: %s", out)
			err = request(config.jiraClient, "DELETE", customFieldContextOptionsDeleteEndpoint(fieldId, contextId, option.Id), nil, nil)
			if err != nil {
				return errors.Wrap(err, "Deleting Jira Field Context Option failed")
			}
		}
	}

	reorderOptsRequest := ReorderFieldOptionsRequest{
		Position:             "First",
		CustomFieldOptionIds: optionIds,
	}

	out, _ := json.Marshal(reorderOptsRequest)
	log.Printf("Sending reorder request: %s", string(out))
	err = request(config.jiraClient, "PUT", customFieldContextOptionsReorderEndpoint(fieldId, contextId), reorderOptsRequest, nil)
	if err != nil {
		return errors.Wrap(err, "Reordering Jira Field Context Options failed")
	}

	return resourceCustomFieldOptionsRead(d, m)
}

// Delete is a no-op, as the concept of an "options list" doesn't actually exist.
// However, as the options depend on the context, and the list depends on the options, it must be separate from the context
func resourceCustomFieldOptionsDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func optionsIdsContains(optionIds []string, option FieldOption) bool {
	for _, id := range optionIds {
		if option.Id == id {
			return true
		}
	}

	return false
}
