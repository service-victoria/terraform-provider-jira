package jira

import (
	"encoding/json"
	"fmt"
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

// FieldRequest The struct sent to the JIRA instance to create a new Field
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

// resourceCustomField is used to define a JIRA custom field's options
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
			"options": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"disabled": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
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
				},
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

	d.Set("options", fromFieldOptions(currentOptions.Values))

	return nil
}

func resourceCustomFieldOptionsUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	opts := d.Get("options").([]interface{})
	options := toFieldOptions(opts)

	newOptions := []FieldOption{}
	existingOptions := []FieldOption{}

	currentOptions := new(GetFieldOptionsResponse)
	err := request(config.jiraClient, "GET", customFieldContextOptionsEndpoint(fieldId, contextId), nil, currentOptions)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Context Options failed")
	}

	for _, option := range options {
		exists, o, _ := optionsContains(currentOptions.Values, option)
		if exists {
			option.Id = o.Id
			existingOptions = append(existingOptions, option)
		} else {
			newOptions = append(newOptions, option)
		}
	}

	for _, option := range currentOptions.Values {
		exists, _, _ := optionsContains(options, option)
		if !exists {
			out, _ := json.Marshal(option)
			log.Printf("Deleting option: %s", out)
			err = request(config.jiraClient, "DELETE", customFieldContextOptionsDeleteEndpoint(fieldId, contextId, option.Id), nil, nil)
			if err != nil {
				return errors.Wrap(err, "Deleting Jira Field Context Option failed")
			}
		}
	}

	updateOptionsResponse := new(FieldOptions)

	if len(existingOptions) > 0 {
		updateOptionsRequest := &FieldOptions{
			Options: existingOptions,
		}
		out, _ := json.Marshal(updateOptionsRequest)
		log.Printf("Updating existing options: %s", out)
		err = request(config.jiraClient, "PUT", customFieldContextOptionsEndpoint(fieldId, contextId), updateOptionsRequest, updateOptionsResponse)
		if err != nil {
			return errors.Wrap(err, "Updating Jira Field Context Options failed")
		}
	}

	createOptionsResponse := new(FieldOptions)

	if len(newOptions) > 0 {
		optionsRequest := &FieldOptions{
			Options: newOptions,
		}
		out, _ := json.Marshal(optionsRequest)
		log.Printf("Creating new options: %s", out)
		err = request(config.jiraClient, "POST", customFieldContextOptionsEndpoint(fieldId, contextId), optionsRequest, createOptionsResponse)
		if err != nil {
			return errors.Wrap(err, "Creating Jira Field Context Options failed")
		}
	}

	updatedOptions := new(GetFieldOptionsResponse)
	err = request(config.jiraClient, "GET", customFieldContextOptionsEndpoint(fieldId, contextId), nil, updatedOptions)
	if err != nil {
		return errors.Wrap(err, "Fetching Jira Field Context Options failed")
	}

	out, _ := json.Marshal(options)
	log.Printf("Options before updates: %s", out)

	for _, option := range updatedOptions.Values {
		out, _ := json.Marshal(option)
		log.Printf("Checking option: %s", out)
		_, _, i := optionsContains(options, option)
		options[i] = option
	}
	out, _ = json.Marshal(options)
	log.Printf("Options after updates: %s", out)

	optionIds := []string{}

	for _, option := range options {
		optionIds = append(optionIds, option.Id)
	}

	reorderOptsRequest := ReorderFieldOptionsRequest{
		Position:             "First",
		CustomFieldOptionIds: optionIds,
	}

	out, _ = json.Marshal(reorderOptsRequest)
	log.Printf("Sending reorder request: %s", string(out))
	err = request(config.jiraClient, "PUT", customFieldContextOptionsReorderEndpoint(fieldId, contextId), reorderOptsRequest, nil)
	if err != nil {
		return errors.Wrap(err, "Reordering Jira Field Context Options failed")
	}

	return resourceCustomFieldOptionsRead(d, m)
}

func resourceCustomFieldOptionsDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	fieldId := d.Get("field_id").(string)
	contextId := d.Get("context_id").(string)

	opts := d.Get("options").([]interface{})
	options := toFieldOptions(opts)

	for _, option := range options {
		exists, _, _ := optionsContains(options, option)
		if !exists {
			err := request(config.jiraClient, "DELETE", customFieldContextOptionsDeleteEndpoint(fieldId, contextId, option.Id), nil, nil)
			if err != nil {
				return errors.Wrap(err, "Deleting Jira Field Context Option failed")
			}
		}
	}

	return nil
}

func optionsContains(opts []FieldOption, opt FieldOption) (bool, FieldOption, int) {
	for i, o := range opts {
		if o.Value == opt.Value || o.Id == opt.Id {
			return true, o, i
		}
	}

	return false, opt, -1
}

func toFieldOptions(opts []interface{}) []FieldOption {
	options := []FieldOption{}
	for _, opt := range opts {
		i := opt.(map[string]interface{})
		o := FieldOption{
			Id:       i["id"].(string),
			Value:    i["value"].(string),
			Disabled: i["disabled"].(bool),
			OptionId: i["option_id"].(string),
		}
		options = append(options, o)
	}
	return options
}

func fromFieldOptions(opts []FieldOption) []interface{} {
	options := make([]interface{}, len(opts), len(opts))
	for i, opt := range opts {
		o := make(map[string]interface{})
		o["id"] = opt.Id
		o["value"] = opt.Value
		o["disabled"] = opt.Disabled
		o["option_id"] = opt.OptionId
		options[i] = o
	}
	return options
}
