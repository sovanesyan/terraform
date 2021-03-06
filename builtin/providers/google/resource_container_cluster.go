package google

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
)

var (
	instanceGroupManagerURL = regexp.MustCompile("^https://www.googleapis.com/compute/v1/projects/([a-z][a-z0-9-]{5}(?:[-a-z0-9]{0,23}[a-z0-9])?)/zones/([a-z0-9-]*)/instanceGroupManagers/([^/]*)")
)

func resourceContainerCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerClusterCreate,
		Read:   resourceContainerClusterRead,
		Update: resourceContainerClusterUpdate,
		Delete: resourceContainerClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"initial_node_count": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"master_auth": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_certificate": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_key": &schema.Schema{
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"cluster_ca_certificate": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": &schema.Schema{
							Type:      schema.TypeString,
							Required:  true,
							ForceNew:  true,
							Sensitive: true,
						},
						"username": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if len(value) > 40 {
						errors = append(errors, fmt.Errorf(
							"%q cannot be longer than 40 characters", k))
					}
					if !regexp.MustCompile("^[a-z0-9-]+$").MatchString(value) {
						errors = append(errors, fmt.Errorf(
							"%q can only contain lowercase letters, numbers and hyphens", k))
					}
					if !regexp.MustCompile("^[a-z]").MatchString(value) {
						errors = append(errors, fmt.Errorf(
							"%q must start with a letter", k))
					}
					if !regexp.MustCompile("[a-z0-9]$").MatchString(value) {
						errors = append(errors, fmt.Errorf(
							"%q must end with a number or a letter", k))
					}
					return
				},
			},

			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"additional_zones": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"cluster_ipv4_cidr": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					_, ipnet, err := net.ParseCIDR(value)

					if err != nil || ipnet == nil || value != ipnet.String() {
						errors = append(errors, fmt.Errorf(
							"%q must contain a valid CIDR", k))
					}
					return
				},
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_group_urls": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"logging_service": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"monitoring_service": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"network": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"subnetwork": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"addons_config": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_load_balancing": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disabled": &schema.Schema{
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"horizontal_pod_autoscaling": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disabled": &schema.Schema{
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"node_config": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"machine_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"disk_size_gb": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(int)

								if value < 10 {
									errors = append(errors, fmt.Errorf(
										"%q cannot be less than 10", k))
								}
								return
							},
						},

						"local_ssd_count": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(int)

								if value < 0 {
									errors = append(errors, fmt.Errorf(
										"%q cannot be negative", k))
								}
								return
							},
						},

						"oauth_scopes": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								StateFunc: func(v interface{}) string {
									return canonicalizeServiceScope(v.(string))
								},
							},
						},

						"service_account": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"metadata": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     schema.TypeString,
						},

						"image_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},

			"node_version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"project": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceContainerClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	zoneName := d.Get("zone").(string)
	clusterName := d.Get("name").(string)

	masterAuths := d.Get("master_auth").([]interface{})
	if len(masterAuths) > 1 {
		return fmt.Errorf("Cannot specify more than one master_auth.")
	}
	masterAuth := masterAuths[0].(map[string]interface{})
	timeoutInMinutes := int(d.Timeout(schema.TimeoutCreate).Minutes())

	cluster := &container.Cluster{
		MasterAuth: &container.MasterAuth{
			Password: masterAuth["password"].(string),
			Username: masterAuth["username"].(string),
		},
		Name:             clusterName,
		InitialNodeCount: int64(d.Get("initial_node_count").(int)),
	}

	if v, ok := d.GetOk("node_version"); ok {
		cluster.InitialClusterVersion = v.(string)
	}

	if v, ok := d.GetOk("additional_zones"); ok {
		locationsList := v.([]interface{})
		locations := []string{}
		for _, v := range locationsList {
			location := v.(string)
			locations = append(locations, location)
			if location == zoneName {
				return fmt.Errorf("additional_zones should not contain the original 'zone'.")
			}
		}
		locations = append(locations, zoneName)
		cluster.Locations = locations
	}

	if v, ok := d.GetOk("cluster_ipv4_cidr"); ok {
		cluster.ClusterIpv4Cidr = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		cluster.Description = v.(string)
	}

	if v, ok := d.GetOk("logging_service"); ok {
		cluster.LoggingService = v.(string)
	}

	if v, ok := d.GetOk("monitoring_service"); ok {
		cluster.MonitoringService = v.(string)
	}

	if _, ok := d.GetOk("network"); ok {
		network, err := getNetworkName(d, "network")
		if err != nil {
			return err
		}
		cluster.Network = network
	}

	if v, ok := d.GetOk("subnetwork"); ok {
		cluster.Subnetwork = v.(string)
	}

	if v, ok := d.GetOk("addons_config"); ok {
		addonsConfig := v.([]interface{})[0].(map[string]interface{})
		cluster.AddonsConfig = &container.AddonsConfig{}

		if v, ok := addonsConfig["http_load_balancing"]; ok {
			addon := v.([]interface{})[0].(map[string]interface{})
			cluster.AddonsConfig.HttpLoadBalancing = &container.HttpLoadBalancing{
				Disabled: addon["disabled"].(bool),
			}
		}

		if v, ok := addonsConfig["horizontal_pod_autoscaling"]; ok {
			addon := v.([]interface{})[0].(map[string]interface{})
			cluster.AddonsConfig.HorizontalPodAutoscaling = &container.HorizontalPodAutoscaling{
				Disabled: addon["disabled"].(bool),
			}
		}
	}
	if v, ok := d.GetOk("node_config"); ok {
		nodeConfigs := v.([]interface{})
		if len(nodeConfigs) > 1 {
			return fmt.Errorf("Cannot specify more than one node_config.")
		}
		nodeConfig := nodeConfigs[0].(map[string]interface{})

		cluster.NodeConfig = &container.NodeConfig{}

		if v, ok = nodeConfig["machine_type"]; ok {
			cluster.NodeConfig.MachineType = v.(string)
		}

		if v, ok = nodeConfig["disk_size_gb"]; ok {
			cluster.NodeConfig.DiskSizeGb = int64(v.(int))
		}

		if v, ok = nodeConfig["local_ssd_count"]; ok {
			cluster.NodeConfig.LocalSsdCount = int64(v.(int))
		}

		if v, ok := nodeConfig["oauth_scopes"]; ok {
			scopesList := v.([]interface{})
			scopes := []string{}
			for _, v := range scopesList {
				scopes = append(scopes, canonicalizeServiceScope(v.(string)))
			}

			cluster.NodeConfig.OauthScopes = scopes
		}

		if v, ok = nodeConfig["service_account"]; ok {
			cluster.NodeConfig.ServiceAccount = v.(string)
		}

		if v, ok = nodeConfig["metadata"]; ok {
			m := make(map[string]string)
			for k, val := range v.(map[string]interface{}) {
				m[k] = val.(string)
			}
			cluster.NodeConfig.Metadata = m
		}

		if v, ok = nodeConfig["image_type"]; ok {
			cluster.NodeConfig.ImageType = v.(string)
		}
	}

	req := &container.CreateClusterRequest{
		Cluster: cluster,
	}

	op, err := config.clientContainer.Projects.Zones.Clusters.Create(
		project, zoneName, req).Do()
	if err != nil {
		return err
	}

	// Wait until it's created
	waitErr := containerOperationWait(config, op, project, zoneName, "creating GKE cluster", timeoutInMinutes, 3)
	if waitErr != nil {
		// The resource didn't actually create
		d.SetId("")
		return waitErr
	}

	log.Printf("[INFO] GKE cluster %s has been created", clusterName)

	d.SetId(clusterName)

	return resourceContainerClusterRead(d, meta)
}

func resourceContainerClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	zoneName := d.Get("zone").(string)

	cluster, err := config.clientContainer.Projects.Zones.Clusters.Get(
		project, zoneName, d.Get("name").(string)).Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			log.Printf("[WARN] Removing Container Cluster %q because it's gone", d.Get("name").(string))
			// The resource doesn't exist anymore
			d.SetId("")

			return nil
		}

		return err
	}

	d.Set("name", cluster.Name)
	d.Set("zone", cluster.Zone)

	locations := []string{}
	if len(cluster.Locations) > 1 {
		for _, location := range cluster.Locations {
			if location != cluster.Zone {
				locations = append(locations, location)
			}
		}
	}
	d.Set("additional_zones", locations)

	d.Set("endpoint", cluster.Endpoint)

	masterAuth := []map[string]interface{}{
		map[string]interface{}{
			"username":               cluster.MasterAuth.Username,
			"password":               cluster.MasterAuth.Password,
			"client_certificate":     cluster.MasterAuth.ClientCertificate,
			"client_key":             cluster.MasterAuth.ClientKey,
			"cluster_ca_certificate": cluster.MasterAuth.ClusterCaCertificate,
		},
	}
	d.Set("master_auth", masterAuth)

	d.Set("initial_node_count", cluster.InitialNodeCount)
	d.Set("node_version", cluster.CurrentNodeVersion)
	d.Set("cluster_ipv4_cidr", cluster.ClusterIpv4Cidr)
	d.Set("description", cluster.Description)
	d.Set("logging_service", cluster.LoggingService)
	d.Set("monitoring_service", cluster.MonitoringService)
	d.Set("network", d.Get("network").(string))
	d.Set("subnetwork", cluster.Subnetwork)
	d.Set("node_config", flattenClusterNodeConfig(cluster.NodeConfig))

	if igUrls, err := getInstanceGroupUrlsFromManagerUrls(config, cluster.InstanceGroupUrls); err != nil {
		return err
	} else {
		d.Set("instance_group_urls", igUrls)
	}

	return nil
}

func resourceContainerClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	zoneName := d.Get("zone").(string)
	clusterName := d.Get("name").(string)
	desiredNodeVersion := d.Get("node_version").(string)
	timeoutInMinutes := int(d.Timeout(schema.TimeoutUpdate).Minutes())

	req := &container.UpdateClusterRequest{
		Update: &container.ClusterUpdate{
			DesiredNodeVersion: desiredNodeVersion,
		},
	}
	op, err := config.clientContainer.Projects.Zones.Clusters.Update(
		project, zoneName, clusterName, req).Do()
	if err != nil {
		return err
	}

	// Wait until it's updated

	waitErr := containerOperationWait(config, op, project, zoneName, "updating GKE cluster", timeoutInMinutes, 2)
	if waitErr != nil {
		return waitErr
	}

	log.Printf("[INFO] GKE cluster %s has been updated to %s", d.Id(),
		desiredNodeVersion)

	return resourceContainerClusterRead(d, meta)
}

func resourceContainerClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	zoneName := d.Get("zone").(string)
	clusterName := d.Get("name").(string)
	timeoutInMinutes := int(d.Timeout(schema.TimeoutDelete).Minutes())

	log.Printf("[DEBUG] Deleting GKE cluster %s", d.Get("name").(string))
	op, err := config.clientContainer.Projects.Zones.Clusters.Delete(
		project, zoneName, clusterName).Do()
	if err != nil {
		return err
	}

	// Wait until it's deleted
	waitErr := containerOperationWait(config, op, project, zoneName, "deleting GKE cluster", timeoutInMinutes, 3)
	if waitErr != nil {
		return waitErr
	}

	log.Printf("[INFO] GKE cluster %s has been deleted", d.Id())

	d.SetId("")

	return nil
}

// container engine's API currently mistakenly returns the instance group manager's
// URL instead of the instance group's URL in its responses. This shim detects that
// error, and corrects it, by fetching the instance group manager URL and retrieving
// the instance group manager, then using that to look up the instance group URL, which
// is then substituted.
//
// This should be removed when the API response is fixed.
func getInstanceGroupUrlsFromManagerUrls(config *Config, igmUrls []string) ([]string, error) {
	instanceGroupURLs := make([]string, 0, len(igmUrls))
	for _, u := range igmUrls {
		if !instanceGroupManagerURL.MatchString(u) {
			instanceGroupURLs = append(instanceGroupURLs, u)
			continue
		}
		matches := instanceGroupManagerURL.FindStringSubmatch(u)
		instanceGroupManager, err := config.clientCompute.InstanceGroupManagers.Get(matches[1], matches[2], matches[3]).Do()
		if err != nil {
			return nil, fmt.Errorf("Error reading instance group manager returned as an instance group URL: %s", err)
		}
		instanceGroupURLs = append(instanceGroupURLs, instanceGroupManager.InstanceGroup)
	}
	return instanceGroupURLs, nil
}

func flattenClusterNodeConfig(c *container.NodeConfig) []map[string]interface{} {
	config := []map[string]interface{}{
		map[string]interface{}{
			"machine_type":    c.MachineType,
			"disk_size_gb":    c.DiskSizeGb,
			"local_ssd_count": c.LocalSsdCount,
			"service_account": c.ServiceAccount,
			"metadata":        c.Metadata,
			"image_type":      c.ImageType,
		},
	}

	if len(c.OauthScopes) > 0 {
		config[0]["oauth_scopes"] = c.OauthScopes
	}

	return config
}
