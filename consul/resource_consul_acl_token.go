package consul

import (
	"fmt"
	"log"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceConsulACLToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceConsulACLTokenCreate,
		Read:   resourceConsulACLTokenRead,
		Update: resourceConsulACLTokenUpdate,
		Delete: resourceConsulACLTokenDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accessor_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Computed:    true,
				Optional:    true,
				Description: "The token id.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The token description.",
			},
			"policies": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of policies.",
			},
			"local": {
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Default:     false,
				Description: "Flag to set the token local to the current datacenter.",
			},
		},
	}
}

func resourceConsulACLTokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := getClient(meta)

	log.Printf("[DEBUG] Creating ACL token")

	aclToken := consulapi.ACLToken{
		AccessorID:  d.Get("accessor_id").(string),
		Description: d.Get("description").(string),
		Local:       d.Get("local").(bool),
	}

	iPolicies := d.Get("policies").(*schema.Set).List()
	policyLinks := make([]*consulapi.ACLTokenPolicyLink, 0, len(iPolicies))
	for _, iPolicy := range iPolicies {
		policyLinks = append(policyLinks, &consulapi.ACLTokenPolicyLink{
			Name: iPolicy.(string),
		})
	}

	if len(policyLinks) > 0 {
		aclToken.Policies = policyLinks
	}

	token, _, err := client.ACL().TokenCreate(&aclToken, nil)
	if err != nil {
		return fmt.Errorf("error creating ACL token: %s", err)
	}

	log.Printf("[DEBUG] Created ACL token %q", token.AccessorID)

	d.SetId(token.AccessorID)

	return resourceConsulACLTokenRead(d, meta)
}

func resourceConsulACLTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := getClient(meta)

	id := d.Id()
	log.Printf("[DEBUG] Reading ACL token %q", id)

	aclToken, _, err := client.ACL().TokenRead(id, nil)
	if err != nil {
		log.Printf("[WARN] ACL token not found, removing from state")
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Read ACL token %q", id)

	if err = d.Set("accessor_id", aclToken.AccessorID); err != nil {
		return fmt.Errorf("Error while setting 'accessor_id': %s", err)
	}

	if err = d.Set("description", aclToken.Description); err != nil {
		return fmt.Errorf("Error while setting 'description': %s", err)
	}

	policies := make([]string, 0, len(aclToken.Policies))
	for _, policyLink := range aclToken.Policies {
		policies = append(policies, policyLink.Name)
	}

	if err = d.Set("policies", policies); err != nil {
		return fmt.Errorf("Error while setting 'policies': %s", err)
	}
	if err = d.Set("local", aclToken.Local); err != nil {
		return fmt.Errorf("Error while setting 'local': %s", err)
	}

	return nil
}

func resourceConsulACLTokenUpdate(d *schema.ResourceData, meta interface{}) error {
	client := getClient(meta)

	id := d.Id()
	log.Printf("[DEBUG] Updating ACL token %q", id)

	aclToken := consulapi.ACLToken{
		AccessorID:  id,
		Description: d.Get("description").(string),
		Local:       d.Get("local").(bool),
	}

	if v, ok := d.GetOk("policies"); ok {
		vs := v.(*schema.Set).List()
		s := make([]*consulapi.ACLTokenPolicyLink, len(vs))
		for i, raw := range vs {
			s[i] = &consulapi.ACLTokenPolicyLink{
				Name: raw.(string),
			}
		}
		aclToken.Policies = s
	}

	_, _, err := client.ACL().TokenUpdate(&aclToken, nil)
	if err != nil {
		return fmt.Errorf("error updating ACL token %q: %s", id, err)
	}
	log.Printf("[DEBUG] Updated ACL token %q", id)

	return resourceConsulACLTokenRead(d, meta)
}

func resourceConsulACLTokenDelete(d *schema.ResourceData, meta interface{}) error {
	client := getClient(meta)

	id := d.Id()

	log.Printf("[DEBUG] Deleting ACL token %q", id)
	_, err := client.ACL().TokenDelete(id, nil)
	if err != nil {
		return fmt.Errorf("error deleting ACL token %q: %s", id, err)
	}
	log.Printf("[DEBUG] Deleted ACL token %q", id)

	return nil
}
