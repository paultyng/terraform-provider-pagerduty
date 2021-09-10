package pagerduty

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

type PagerDutyExtensionServiceNowConfig struct {
	User        string `json:"snow_user"`
	Password    string `json:"snow_password,omitempty"`
	SyncOptions string `json:"sync_options"`
	Target      string `json:"target"`
	TaskType    string `json:"task_type"`
	Referer     string `json:"referer"`
}

func resourcePagerDutyExtensionServiceNow() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyExtensionServiceNowCreate,
		Read:   resourcePagerDutyExtensionServiceNowRead,
		Update: resourcePagerDutyExtensionServiceNowUpdate,
		Delete: resourcePagerDutyExtensionServiceNowDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyExtensionServiceNowImport,
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
			"snow_user": {
				Type:     schema.TypeString,
				Required: true,
			},
			"snow_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"sync_options": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"manual_sync", "sync_all"}, false),
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
			},
			"task_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"referer": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func buildExtensionServiceNowStruct(d *schema.ResourceData) *pagerduty.Extension {
	Extension := &pagerduty.Extension{
		Name:        d.Get("name").(string),
		Type:        "extension",
		EndpointURL: d.Get("endpoint_url").(string),
		ExtensionSchema: &pagerduty.ExtensionSchemaReference{
			Type: "extension_schema_reference",
			ID:   d.Get("extension_schema").(string),
		},
		ExtensionObjects: expandServiceNowServiceObjects(d.Get("extension_objects")),
	}

	var config = &PagerDutyExtensionServiceNowConfig{
		User:        d.Get("snow_user").(string),
		Password:    d.Get("snow_password").(string),
		SyncOptions: d.Get("sync_options").(string),
		Target:      d.Get("target").(string),
		TaskType:    d.Get("task_type").(string),
		Referer:     d.Get("referer").(string),
	}
	Extension.Config = config

	return Extension
}

func resourcePagerDutyExtensionServiceNowCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	extension := buildExtensionServiceNowStruct(d)

	log.Printf("[INFO] Creating PagerDuty extension %s", extension.Name)

	extension, _, err := client.Extensions.Create(extension)
	if err != nil {
		return err
	}

	d.SetId(extension.ID)

	return resourcePagerDutyExtensionServiceNowRead(d, meta)
}

func resourcePagerDutyExtensionServiceNowRead(d *schema.ResourceData, meta interface{}) error {
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
		if err := d.Set("extension_objects", flattenExtensionServiceNowObjects(extension.ExtensionObjects)); err != nil {
			log.Printf("[WARN] error setting extension_objects: %s", err)
		}
		d.Set("extension_schema", extension.ExtensionSchema.ID)

		b, _ := json.Marshal(extension.Config)
		var config = new(PagerDutyExtensionServiceNowConfig)
		json.Unmarshal(b, config)
		d.Set("snow_user", config.User)
		d.Set("snow_password", config.Password)
		d.Set("sync_options", config.SyncOptions)
		d.Set("target", config.Target)
		d.Set("task_type", config.TaskType)
		d.Set("referer", config.Referer)

		return nil
	})
}

func resourcePagerDutyExtensionServiceNowUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	extension := buildExtensionServiceNowStruct(d)

	log.Printf("[INFO] Updating PagerDuty extension %s", d.Id())

	if _, _, err := client.Extensions.Update(d.Id(), extension); err != nil {
		return err
	}

	return resourcePagerDutyExtensionServiceNowRead(d, meta)
}

func resourcePagerDutyExtensionServiceNowDelete(d *schema.ResourceData, meta interface{}) error {
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

func resourcePagerDutyExtensionServiceNowImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func expandServiceNowServiceObjects(v interface{}) []*pagerduty.ServiceReference {
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

func flattenExtensionServiceNowObjects(serviceList []*pagerduty.ServiceReference) interface{} {
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
