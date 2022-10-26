package wallarm

import (
	"encoding/json"
	"fmt"
	"log"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmGlobalMode() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmGlobalModeCreate,
		Read:   resourceWallarmGlobalModeRead,
		Update: resourceWallarmGlobalModeUpdate,
		Delete: resourceWallarmGlobalModeDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Client ID to perform changes",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},

			"filtration_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "default",
				ValidateFunc: validation.StringInSlice([]string{"default", "monitoring", "block", "safe_blocking", "off"}, false),
			},

			"scanner_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "on",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},

			"rechecker_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "off",
				ValidateFunc: validation.StringInSlice([]string{"on", "off"}, false),
			},
		},
	}
}

func resourceWallarmGlobalModeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

	filtrationMode := d.Get("filtration_mode").(string)

	_, err := client.WallarmModeUpdate(&wallarm.WallarmModeParams{Mode: filtrationMode}, clientID)
	if err != nil {
		return err
	}

	scannerMode := d.Get("scanner_mode").(string)
	if scannerMode == "on" {
		scannerMode = "classic"
	}

	recheckerMode := d.Get("rechecker_mode").(string)

	mode := &wallarm.ClientUpdate{
		Filter: &wallarm.ClientFilter{
			ID: clientID,
		},
		Fields: &wallarm.ClientFields{
			ScannerMode:         scannerMode,
			AttackRecheckerMode: recheckerMode,
		},
	}
	_, err = client.ClientUpdate(mode)
	if err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%s/%s", clientID, filtrationMode, scannerMode, recheckerMode)
	d.SetId(resID)

	d.Set("client_id", clientID)

	return resourceWallarmGlobalModeRead(d, m)
}

func resourceWallarmGlobalModeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)

	wallarmModeResp, err := client.WallarmModeRead(clientID)
	if err != nil {
		return err
	}
	if wallarmModeResp.Status != 200 {
		body, err := json.Marshal(wallarmModeResp)
		if err != nil {
			return err
		}
		log.Printf("[WARN] Couldn't fetch wallarm_mode. Body: %s", body)

		d.SetId("")
		return nil
	}

	filtrationMode := wallarmModeResp.Body.Mode
	if err := d.Set("filtration_mode", filtrationMode); err != nil {
		return err
	}

	clientInfo := &wallarm.ClientRead{
		Filter: &wallarm.ClientReadFilter{
			Enabled: true,
			ClientFilter: wallarm.ClientFilter{
				ID: clientID},
		},
		Limit:  1000,
		Offset: 0,
	}

	otherModesResp, err := client.ClientRead(clientInfo)
	if err != nil {
		return err
	}
	if len(otherModesResp.Body) == 0 {
		body, err := json.Marshal(otherModesResp)
		if err != nil {
			return err
		}
		log.Printf("[WARN] Client hasn't been found in API. Body: %s", body)

		d.SetId("")
		return nil
	}

	scannerMode := otherModesResp.Body[0].ScannerMode
	if scannerMode == "classic" {
		scannerMode = "on"
	}

	if err := d.Set("scanner_mode", scannerMode); err != nil {
		return err
	}

	recheckerMode := otherModesResp.Body[0].AttackRecheckerMode

	if err := d.Set("rechecker_mode", recheckerMode); err != nil {
		return err
	}

	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmGlobalModeUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceWallarmGlobalModeCreate(d, m)
}

func resourceWallarmGlobalModeDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
