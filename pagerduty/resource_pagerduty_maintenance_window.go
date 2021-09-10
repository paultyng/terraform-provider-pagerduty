package pagerduty

import (
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyMaintenanceWindowCreate,
		Read:   resourcePagerDutyMaintenanceWindowRead,
		Update: resourcePagerDutyMaintenanceWindowUpdate,
		Delete: resourcePagerDutyMaintenanceWindowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"start_time": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateRFC3339,
				DiffSuppressFunc: suppressRFC3339Diff,
			},
			"end_time": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateRFC3339,
				DiffSuppressFunc: suppressRFC3339Diff,
			},

			"services": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
		},
	}
}

func buildMaintenanceWindowStruct(d *schema.ResourceData) *pagerduty.MaintenanceWindow {
	window := &pagerduty.MaintenanceWindow{
		StartTime: d.Get("start_time").(string),
		EndTime:   d.Get("end_time").(string),
		Services:  expandServices(d.Get("services").(*schema.Set)),
	}

	if v, ok := d.GetOk("description"); ok {
		window.Description = v.(string)
	}

	return window
}

func resourcePagerDutyMaintenanceWindowCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	window := buildMaintenanceWindowStruct(d)

	log.Printf("[INFO] Creating PagerDuty maintenance window")

	window, _, err := client.MaintenanceWindows.Create(window)
	if err != nil {
		return err
	}

	d.SetId(window.ID)

	return nil
}

func resourcePagerDutyMaintenanceWindowRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty maintenance window %s", d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		window, _, err := client.MaintenanceWindows.Get(d.Id())
		if err != nil {
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}

		d.Set("description", window.Description)
		d.Set("start_time", window.StartTime)
		d.Set("end_time", window.EndTime)

		if err := d.Set("services", flattenServices(window.Services)); err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
}

func resourcePagerDutyMaintenanceWindowUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	window := buildMaintenanceWindowStruct(d)

	log.Printf("[INFO] Updating PagerDuty maintenance window %s", d.Id())

	if _, _, err := client.MaintenanceWindows.Update(d.Id(), window); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyMaintenanceWindowDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty maintenance window %s", d.Id())

	if _, err := client.MaintenanceWindows.Delete(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func expandServices(v *schema.Set) []*pagerduty.ServiceReference {
	var services []*pagerduty.ServiceReference

	for _, srv := range v.List() {
		service := &pagerduty.ServiceReference{
			Type: "service_reference",
			ID:   srv.(string),
		}
		services = append(services, service)
	}

	return services
}

func flattenServices(v []*pagerduty.ServiceReference) *schema.Set {
	var services []interface{}

	for _, srv := range v {
		services = append(services, srv.ID)
	}

	return schema.NewSet(schema.HashString, services)
}
