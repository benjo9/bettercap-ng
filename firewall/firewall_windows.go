package firewall

import (
	"fmt"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/net"
)

type WindowsFirewall struct {
	iface        *net.Endpoint
	forwarding   bool
	redirections map[string]*Redirection
}

func Make(iface *net.Endpoint) FirewallManager {
	firewall := &WindowsFirewall{
		iface:        iface,
		forwarding:   false,
		redirections: make(map[string]*Redirection, 0),
	}

	firewall.forwarding = firewall.IsForwardingEnabled()

	return firewall
}

func (f WindowsFirewall) IsForwardingEnabled() bool {
	if out, err := core.Exec("netsh", []string{"interface", "ipv4", "dump"}); err != nil {
		fmt.Printf("%s\n", err)
		return false
	} else {
		return strings.Contains(out, "forwarding=enabled")
	}
}

func (f WindowsFirewall) EnableForwarding(enabled bool) error {
	v := "enabled"
	if enabled == false {
		v = "disabled"
	}
	out, err := core.Exec("netsh", []string{"interface", "ipv4", "set", "interface", fmt.Sprintf("%d", f.iface.Index), fmt.Sprintf("forwarding=\"%s\"", v)})
	if err != nil {
		return err
	}

	if strings.Contains(out, "OK") == false {
		return fmt.Errorf("Unexpected netsh output: %s", out)
	}

	return nil
}

func (f WindowsFirewall) generateRule(r *Redirection, enabled bool) []string {
	rule := []string{
		"listenaddress=0.0.0.0",
		fmt.Sprintf("listenport=%d", r.SrcPort),
	}

	if enabled == true {
		rule = append(rule, fmt.Sprintf("protocol=%s", r.Protocol))
		rule = append(rule, fmt.Sprintf("connectport=%d", r.DstPort))
		rule = append(rule, fmt.Sprintf("connectaddress=%s", r.DstAddress))
	}

	return rule
}

func (f *WindowsFirewall) EnableRedirection(r *Redirection, enabled bool) error {
	fmt.Printf("%s: Port redirection is not supported on Windows, packets will NOT be automatically forwarded to the proxy.", core.Yellow("WARNING"))
	/*
		TODO: This doesn't work :/

		rule := f.generateRule(r, enabled)
		if enabled == true {
			rule = append([]string{"interface", "portproxy", "add", "v4tov4"}, rule...)
		} else {
			rule = append([]string{"interface", "portproxy", "delete", "v4tov4"}, rule...)
		}

		out, err := core.Exec("netsh", rule)
		if err != nil {
			return err
		}

		if core.Trim(out) != "" {
			return fmt.Errorf("Unexpected netsh output: %s", out)
		}
	*/
	return nil
}

func (f WindowsFirewall) Restore() {
	for _, r := range f.redirections {
		if err := f.EnableRedirection(r, false); err != nil {
			fmt.Printf("%s", err)
		}
	}

	if err := f.EnableForwarding(f.forwarding); err != nil {
		fmt.Printf("%s", err)
	}
}
