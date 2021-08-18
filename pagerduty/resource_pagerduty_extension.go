package pagerduty

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyExtension() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyExtensionCreate,
		Read:   resourcePagerDutyExtensionRead,
		Update: resourcePagerDutyExtensionUpdate,
		Delete: resourcePagerDutyExtensionDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyExtensionImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"endpoint_url": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"extension_objects": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"extension_schema": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
			"summary": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func buildExtensionStruct(d *schema.ResourceData) *pagerduty.Extension {
	Extension := &pagerduty.Extension{
		Name:        d.Get("name").(string),
		Type:        "extension",
		EndpointURL: d.Get("endpoint_url").(string),
		ExtensionSchema: &pagerduty.ExtensionSchemaReference{
			Type: "extension_schema_reference",
			ID:   d.Get("extension_schema").(string),
		},
		ExtensionObjects: expandServiceObjects(d.Get("extension_objects")),
	}

	if v, ok := d.GetOk("config"); ok {
		Extension.Config = expandExtensionConfig(v)
	}

	return Extension
}

func resourcePagerDutyExtensionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	extension := buildExtensionStruct(d)

	log.Printf("[INFO] Creating PagerDuty extension %s", extension.Name)

	extension, _, err := client.Extensions.Create(extension)
	if err != nil {
		return err
	}

	d.SetId(extension.ID)

	return resourcePagerDutyExtensionRead(d, meta)
}

func resourcePagerDutyExtensionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty extension %s", d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		extension, _, err := client.Extensions.Get(d.Id())
		if err != nil {
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}

		d.Set("summary", extension.Summary)
		d.Set("name", extension.Name)
		d.Set("endpoint_url", extension.EndpointURL)
		d.Set("html_url", extension.HTMLURL)
		if err := d.Set("extension_objects", flattenExtensionObjects(extension.ExtensionObjects)); err != nil {
			log.Printf("[WARN] error setting extension_objects: %s", err)
		}
		d.Set("extension_schema", extension.ExtensionSchema.ID)

		if err := d.Set("config", flattenExtensionConfig(extension.Config)); err != nil {
			log.Printf("[WARN] error setting extension config: %s", err)
		}

		return nil
	})
}

func resourcePagerDutyExtensionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	extension := buildExtensionStruct(d)

	log.Printf("[INFO] Updating PagerDuty extension %s", d.Id())

	if _, _, err := client.Extensions.Update(d.Id(), extension); err != nil {
		return err
	}

	return resourcePagerDutyExtensionRead(d, meta)
}

func resourcePagerDutyExtensionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty extension %s", d.Id())

	if _, err := client.Extensions.Delete(d.Id()); err != nil {
		if perr, ok := err.(*pagerduty.Error); ok && perr.Code == 5001 {
			log.Printf("[WARN] Extension (%s) not found, removing from state", d.Id())
			return nil
		}
		return err
	}

	d.SetId("")

	return nil
}

func resourcePagerDutyExtensionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	extension, _, err := client.Extensions.Get(d.Id())

	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("error importing pagerduty_extension. Expecting an importation ID for extension")
	}

	d.Set("endpoint_url", extension.EndpointURL)
	d.Set("extension_objects", []string{extension.ExtensionObjects[0].ID})
	d.Set("extension_schema", extension.ExtensionSchema.ID)

	return []*schema.ResourceData{d}, err
}

func expandServiceObjects(v interface{}) []*pagerduty.ServiceReference {
	var services []*pagerduty.ServiceReference

	for _, srv := range v.(*schema.Set).List() {
		service := &pagerduty.ServiceReference{
			Type: "service_reference",
			ID:   srv.(string),
		}
		services = append(services, service)
	}

	return services
}

func flattenExtensionObjects(serviceList []*pagerduty.ServiceReference) interface{} {
	var services []interface{}
	for _, s := range serviceList {
		// only flatten service_reference types, because that's all we send at this
		// time
		if s.Type == "service_reference" {
			services = append(services, s.ID)
		}
	}
	return services
}
func expandExtensionConfig(v interface{}) interface{} {
	var config interface{}
	if err := json.Unmarshal([]byte(v.(string)), &config); err != nil {
		log.Printf("[ERROR] Could not unmarshal extension config %s: %v", v.(string), err)
		return nil
	}

	return config
}

func flattenExtensionConfig(config interface{}) interface{} {
	json, err := json.Marshal(config)
	if err != nil {
		log.Printf("[ERROR] Could not marshal extension config %s: %v", config.(string), err)
		return nil
	}
	return string(json)
}
