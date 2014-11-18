package xenserver

import (
    "github.com/nilshell/xmlrpc"
    "log"
    "fmt"
    "errors"
)

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}


type XenAPIClient struct {
    Session     interface{}
    Host        string
    Url         string
    Username    string
    Password    string
    RPC         *xmlrpc.Client
}


type APIResult struct {
    Status string
    Value interface{}
    ErrorDescription string
}

type VM struct {
    Ref     string
    Client  *XenAPIClient
}

type SR struct {
    Ref     string
    Client  *XenAPIClient
}

type VDI struct {
    Ref     string
    Client  *XenAPIClient
}

type Network struct {
    Ref     string
    Client  *XenAPIClient
}

type VBD struct {
    Ref string
    Client *XenAPIClient
}

type VIF struct {
    Ref string
    Client *XenAPIClient
}

func (c *XenAPIClient) RPCCall (result interface{}, method string, params []interface{}) (err error) {
    fmt.Println(params)
    p := new(xmlrpc.Params)
    p.Params = params
    err = c.RPC.Call(method, *p, result)
    return err
}


func (client *XenAPIClient) Login () (err error) {
    //Do loging call
    result := xmlrpc.Struct{}

    params := make([]interface{}, 2)
    params[0] = client.Username
    params[1] = client.Password

    err = client.RPCCall(&result, "session.login_with_password", params)
    client.Session = result["Value"]
    return err
}

func (client *XenAPIClient) APICall (result *APIResult, method string, params ...interface{}) (err error) {
    if client.Session == nil {
        fmt.Println("Error: no session")
        return fmt.Errorf("No session. Unable to make call")
    }

    //Make a params slice which will include the session
    p := make([]interface{}, len(params) + 1)
    p[0] = client.Session

    if params != nil {
        for idx, element := range params {
            p[idx+1] = element
        }
    }

    res := xmlrpc.Struct{}

    err = client.RPCCall(&res, method, p)

    if err != nil {
        return err
    }

    result.Status = res["Status"].(string)

    if result.Status != "Success" {
        fmt.Println("Encountered an API error: ", result.Status)
        fmt.Println(res["ErrorDescription"])
        log.Fatal(res["ErrorDescription"])
        return errors.New("API Error occurred")
    } else {
        result.Value = res["Value"]
    }
    return
}


func (client *XenAPIClient) GetHosts () (err error) {
    result := APIResult{}
    _ = client.APICall(&result, "host.get_all")
    hosts := result.Value
    fmt.Println(hosts)
    return nil
}


func (client *XenAPIClient) GetVMByUuid (vm_uuid string) (vm *VM, err error) {
    vm = new(VM)
    result := APIResult{}
    err = client.APICall(&result, "VM.get_by_uuid", vm_uuid)
    if err != nil {
        return nil, err
    }
    vm.Ref = result.Value.(string)
    vm.Client = client
    return
}

func (client *XenAPIClient) GetNetworkByUuid (network_uuid string) (network *Network, err error) {
    network = new(Network)
    result := APIResult{}
    err = client.APICall(&result, "network.get_by_uuid", network_uuid)
    if err != nil {
        return nil, err
    }
    network.Ref = result.Value.(string)
    network.Client = client
    return
}


func (client *XenAPIClient) GetNetworkByNameLabel (name_label string) (networks []*Network, err error) {
    networks = make([]*Network, 0)
    result := APIResult{}
    err = client.APICall(&result, "network.get_by_name_label", name_label)
    if err != nil {
        return networks, err
    }

    for _, elem := range result.Value.([]interface{}) {
        network := new(Network)
        network.Ref = elem.(string)
        network.Client = client
        networks = append(networks, network)
    }

    return networks, nil
}


func (client *XenAPIClient) GetSRByUuid (sr_uuid string) (sr *SR, err error) {
    sr = new(SR)
    result := APIResult{}
    err = client.APICall(&result, "SR.get_by_uuid", sr_uuid)
    if err != nil {
        return nil, err
    }
    sr.Ref = result.Value.(string)
    sr.Client = client
    return
}

func (client *XenAPIClient) GetVdiByUuid (vdi_uuid string) (vdi *VDI, err error) {
    vdi = new(VDI)
    result := APIResult{}
    err = client.APICall(&result, "VDI.get_by_uuid", vdi_uuid)
    if err != nil {
        return nil, err
    }
    vdi.Ref = result.Value.(string)
    vdi.Client = client
    return
}


// VM associated functions

func (self *VM) Clone (label string) (new_instance *VM, err error) {
    new_instance = new(VM)

    result := APIResult{}
    err = self.Client.APICall(&result, "VM.clone", self.Ref, label)
    if err != nil {
        return nil, err
    }
    new_instance.Ref = result.Value.(string)
    new_instance.Client = self.Client
    return
}

func (self *VM) Start(paused, force bool) (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.start", self.Ref, paused, force)
    if err != nil {
        return err
    }
    return
}

func (self *VM) CleanShutdown() (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.clean_shutdown", self.Ref)
    if err != nil {
        return err
    }
    return
}

func (self *VM) Unpause () (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.unpause", self.Ref)
    if err != nil {
        return err
    }
    return
}

func (self *VM) SetPVBootloader(pv_bootloader, pv_args string) (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.set_PV_bootloader", self.Ref, pv_bootloader)
    if err != nil {
        return err
    }
    result = APIResult{}
    err = self.Client.APICall(&result, "VM.set_PV_bootloader_args", self.Ref, pv_args)
    if err != nil {
        return err
    }
    return
}

func (self *VM) GetDomainId() (domid string, err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.get_domid", self.Ref)
    if err != nil {
        return "", err
    }
    domid = result.Value.(string)
    return domid, nil
}

func (self *VM) GetPowerState() (state string, err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.get_power_state", self.Ref)
    if err != nil {
        return "", err
    }
    state = result.Value.(string)
    return state, nil
}

func (self *VM) GetUuid() (uuid string, err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.get_uuid", self.Ref)
    if err != nil {
        return "", err
    }
    uuid = result.Value.(string)
    return uuid, nil
}

func (self *VM) GetVBDs() (vbds []VBD, err error) {
    vbds = make([]VBD, 0)
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.get_VBDs", self.Ref)
    if err != nil {
        return vbds, err
    }
    for _, elem := range result.Value.([]interface{}) {
        vbd := VBD{}
        vbd.Ref = elem.(string)
        vbd.Client = self.Client
        vbds = append(vbds, vbd)
    }

    return vbds, nil
}


func (self *VM) GetVIFs() (vifs []VIF, err error) {
    vifs = make([]VIF, 0)
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.get_VIFs", self.Ref)
    if err != nil {
        return vifs, err
    }
    for _, elem := range result.Value.([]interface{}) {
        vif := VIF{}
        vif.Ref = elem.(string)
        vif.Client = self.Client
        vifs = append(vifs, vif)
    }

    return vifs, nil
}

func (self *VM) GetDisks() (vdis []*VDI, err error) {
    // Return just data disks (non-isos)
    vdis = make([]*VDI, 0)
    vbds, err := self.GetVBDs()
    if err != nil {
        return nil, err
    }

    for _, vbd := range vbds {
        rec, err := vbd.GetRecord()
        if err != nil {
            return nil, err
        }
        if rec["type"] == "Disk" {

            vdi, err := vbd.GetVDI()
            if err != nil {
                return nil, err
            }
            vdis = append(vdis, vdi)

        }
    }
    return vdis, nil
}

func (self *VM) GetGuestMetricsRef() (ref string, err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.get_guest_metrics", self.Ref)
    if err != nil {
        return "", nil
    }
    ref = result.Value.(string)
    return ref, err
}

func (self *VM) GetGuestMetrics() (metrics map[string]interface{}, err error) {
    metrics = make(map[string]interface{})
    metrics_ref, err := self.GetGuestMetricsRef()
    if err != nil {
        return metrics, err
    }

    result := APIResult{}
    err = self.Client.APICall(&result, "VM_guest_metrics.get_record", metrics_ref)
    if err != nil {
        return metrics, nil
    }
    for k, v := range result.Value.(xmlrpc.Struct) {
        metrics[k] = v
    }
    return metrics, nil
}

func (self *VM) SetStaticMemoryRange(min, max string) (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.set_memory_limits", self.Ref, min, max, min, max)
    if err != nil {
        return err
    }
    return
}

func (self *VM) ConnectVdi (vdi *VDI, iso bool) (err error) {

    // 1. Create a VBD

    vbd_rec := make(xmlrpc.Struct)
    vbd_rec["VM"] = self.Ref
    vbd_rec["VDI"] = vdi.Ref
    vbd_rec["userdevice"] = "autodetect"
    vbd_rec["unpluggable"] = false
    vbd_rec["empty"] = false
    vbd_rec["other_config"] = make(xmlrpc.Struct)
    vbd_rec["qos_algorithm_type"] = ""
    vbd_rec["qos_algorithm_params"] = make(xmlrpc.Struct)

    if iso {
        vbd_rec["mode"] = "RO"
        vbd_rec["bootable"] = true
        vbd_rec["type"] = "CD"
    } else {
        vbd_rec["mode"] = "RW"
        vbd_rec["bootable"] = false
        vbd_rec["type"] = "Disk"
    }

    result := APIResult{}
    err = self.Client.APICall(&result, "VBD.create", vbd_rec)

    if err != nil {
        return err
    }

    vbd_ref := result.Value.(string)
    fmt.Println("VBD Ref:", vbd_ref)

    result = APIResult{}
    err = self.Client.APICall(&result, "VBD.get_uuid", vbd_ref) 

    fmt.Println("VBD UUID: ", result.Value.(string))
/*
    // 2. Plug VBD (Non need - the VM hasn't booted.
    // @todo - check VM state
    result = APIResult{}
    err = self.Client.APICall(&result, "VBD.plug", vbd_ref)

    if err != nil {
        return err
    }
*/
    return
}

func (self *VM) SetPlatform(params map[string]string) (err error) {
    result := APIResult{}
    platform_rec := make(xmlrpc.Struct)
    for key, value := range params {
        platform_rec[key] = value
    }

    err = self.Client.APICall(&result, "VM.set_platform", self.Ref, platform_rec)

    if err != nil {
        return err
    }
    return
}


func (self *VM) ConnectNetwork (network *Network, device string) (vif *VIF, err error) {
    // Create the VIF

    vif_rec := make(xmlrpc.Struct)
    vif_rec["network"] = network.Ref
    vif_rec["VM"] = self.Ref
    vif_rec["MAC"] = ""
    vif_rec["device"] = device
    vif_rec["MTU"] = "1504"
    vif_rec["other_config"] = make(xmlrpc.Struct)
    vif_rec["qos_algorithm_type"] = ""
    vif_rec["qos_algorithm_params"] = make(xmlrpc.Struct)

    result := APIResult{}
    err = self.Client.APICall(&result, "VIF.create", vif_rec)

    if err != nil {
        return nil, err
    }

    vif = new(VIF)
    vif.Ref = result.Value.(string)
    vif.Client = self.Client

    return vif, nil
}

//      Setters

func (self *VM) SetIsATemplate (is_a_template bool) (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VM.set_is_a_template", self.Ref, is_a_template)
    if err != nil {
        return err
    }
    return
}

// SR associated functions

func (self *SR) CreateVdi (name_label, size string) (vdi *VDI, err error) {
    vdi = new(VDI)

    vdi_rec := make(xmlrpc.Struct)
    vdi_rec["name_label"] = name_label
    vdi_rec["SR"] = self.Ref
    vdi_rec["virtual_size"] = size
    vdi_rec["type"] = "user"
    vdi_rec["sharable"] = false
    vdi_rec["read_only"] = false

    oc := make(xmlrpc.Struct)
    oc["temp"] = "temp"
    vdi_rec["other_config"] = oc


    result := APIResult{}
    err = self.Client.APICall(&result, "VDI.create", vdi_rec)
    if err != nil {
        return nil, err
    }

    vdi.Ref = result.Value.(string)
    vdi.Client = self.Client

    return
}

// Network associated functions

func (self *Network) GetAssignedIPs () (ip_map map[string]string, err error) {
    ip_map = make(map[string]string, 0)
    result := APIResult{}
    err = self.Client.APICall(&result, "network.get_assigned_ips", self.Ref)
    if err != nil {
        return ip_map, err
    }
    for k, v := range result.Value.(xmlrpc.Struct) {
        ip_map[k] = v.(string)
    }
    return ip_map, nil
}

// VBD associated functions
func (self *VBD) GetRecord () (record map[string]interface{}, err error) {
    record = make(map[string]interface{})
    result := APIResult{}
    err = self.Client.APICall(&result, "VBD.get_record", self.Ref)
    if err != nil {
        return record, err
    }
    for k, v := range result.Value.(xmlrpc.Struct) {
        record[k] = v
    }
    return record, nil
}

func (self *VBD) GetVDI () (vdi *VDI, err error) {
    vbd_rec, err := self.GetRecord()
    if err != nil {
        return nil, err
    }

    vdi = new(VDI)
    vdi.Ref = vbd_rec["VDI"].(string)
    vdi.Client = self.Client

    return vdi, nil
}

func (self *VBD) Eject () (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VBD.eject", self.Ref)
    if err != nil {
        return err
    }
    return nil
}

// VIF associated functions

func (self *VIF) Destroy () (err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VIF.destroy", self.Ref)
    if err != nil {
        return err
    }
    return nil
}

// VDI associated functions

func (self *VDI) GetUuid () (vdi_uuid string, err error) {
    result := APIResult{}
    err = self.Client.APICall(&result, "VDI.get_uuid", self.Ref)
    if err != nil {
        return "", err
    }
    vdi_uuid = result.Value.(string)
    return vdi_uuid, nil
}


// Client Initiator

func NewXenAPIClient (host, username, password string) (client XenAPIClient) {
    client.Host = host
    client.Url = "http://" + host
    client.Username = username
    client.Password = password
    client.RPC, _ = xmlrpc.NewClient(client.Url, nil)
    return
}