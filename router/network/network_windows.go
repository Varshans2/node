// +build windows

/*
 * Copyright (C) 2020 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package network

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/jackpal/gateway"
)

// RoutingTable implements a set of platform specific tool for creating, deleting
// and observe routing tables rules for a different needs.
type RoutingTable struct{}

// DiscoverGateway returns system default gateway.
func (t *RoutingTable) DiscoverGateway() (net.IP, error) {
	return gateway.DiscoverGateway()
}

// ExcludeRule adds a single IP address to be excluded from the main tunnelled traffic.
// Traffic sent to the IP address will be directed to the system default gaitway
// instead of tunnel.
func (t *RoutingTable) ExcludeRule(ip, gw net.IP) error {
	out, err := exec.Command("powershell", "-Command", "route add "+ip.String()+"/32 "+gw.String()).CombinedOutput()
	return fmt.Errorf("%s: %w", string(out), err)
}

// DeleteRule removes excluded routing table rule to return it back to routing
// thought the tunnel.
func (t *RoutingTable) DeleteRule(ip, gw net.IP) error {
	out, err := exec.Command("powershell", "-Command", "route delete "+ip.String()+"/32").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete route: %w, %s", err, string(out))
	}

	return nil
}
