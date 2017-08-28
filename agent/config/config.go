package config

import (
	"flag"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/hcl"
)

// Flags defines the command line flags.
//
// All fields are specified as pointers to simplify merging multiple
// File structures since this allows to determine whether a field has
// been set.
type Flags struct {
	File
	ConfigFiles []string
}

// ParseFlag parses the arguments into a Flags struct.
func ParseFlags(args []string) (Flags, error) {
	var f Flags
	fs := NewFlagSet(&f)
	if err := fs.Parse(args); err != nil {
		return Flags{}, err
	}
	return f, nil
}

// NewFlagSet creates the set of command line flags for the agent.
func NewFlagSet(f *Flags) *flag.FlagSet {
	fs := FlagSet{flag.NewFlagSet("agent", flag.ContinueOnError)}

	fs.StringPtrVar(&f.BindAddr, "bind", "bind address")
	fs.BoolPtrVar(&f.Bootstrap, "bootstrap", "bootstrap yes/no")
	fs.StringSliceVar(&f.ConfigFiles, "config-dir", "config dir")
	fs.StringSliceVar(&f.ConfigFiles, "config-file", "config file")
	fs.StringPtrVar(&f.Datacenter, "datacenter", "datacenter")
	fs.IntPtrVar(&f.Ports.DNS, "dns-port", "DNS port")
	fs.StringSliceVar(&f.JoinAddrsLAN, "join", "join addrs")
	fs.StringMapVar(&f.NodeMeta, "node-meta", "node meta as key:val")

	return fs.FlagSet
}

// File defines the format of a config file.
//
// All fields are specified as pointers to simplify merging multiple
// File structures since this allows to determine whether a field has
// been set.
type File struct {
	Bootstrap           *bool
	CheckUpdateInterval *string `json:"check_update_interval" hcl:"check_update_interval"`
	Datacenter          *string
	BindAddr            *string           `json:"bind_addr" hcl:"bind_addr"`
	JoinAddrsLAN        []string          `json:"start_join" hcl:"start_join"`
	NodeMeta            map[string]string `json:"node_meta" hcl:"node_meta"`
	Ports               FilePorts
}

type FilePorts struct {
	DNS *int
}

// ParseFile decodes a configuration file in JSON or HCL format.
func ParseFile(s string) (File, error) {
	var f File
	if err := hcl.Decode(&f, s); err != nil {
		return File{}, err
	}
	return f, nil
}

// Config is the runtime configuration.
type Config struct {
	// simple values

	Bootstrap           bool
	CheckUpdateInterval time.Duration
	Datacenter          string

	// address values

	BindAddrs    []string
	JoinAddrsLAN []string

	// server endpoint values

	DNSPort     int
	DNSAddrsTCP []string
	DNSAddrsUDP []string

	// other values

	NodeMeta map[string]string
}

// NewConfig creates the runtime configuration from a configuration
// file. It performs all the necessary syntactic and semantic validation
// so that the resulting runtime configuration is usable.
func NewConfig(f File) (c Config, err error) {
	boolVal := func(b *bool) bool {
		if err != nil || b == nil {
			return false
		}
		return *b
	}

	durationVal := func(s *string) (d time.Duration) {
		if err != nil || s == nil {
			return 0
		}
		d, err = time.ParseDuration(*s)
		return
	}

	intVal := func(n *int) int {
		if err != nil || n == nil {
			return 0
		}
		return *n
	}

	stringVal := func(s *string) string {
		if err != nil || s == nil {
			return ""
		}
		return *s
	}

	addrVal := func(s *string) string {
		addr := stringVal(s)
		if addr == "" {
			return "0.0.0.0"
		}
		return addr
	}

	joinHostPort := func(host string, port int) string {
		if host == "0.0.0.0" {
			host = ""
		}
		return net.JoinHostPort(host, strconv.Itoa(port))
	}

	c.Bootstrap = boolVal(f.Bootstrap)
	c.CheckUpdateInterval = durationVal(f.CheckUpdateInterval)
	c.Datacenter = stringVal(f.Datacenter)
	c.JoinAddrsLAN = f.JoinAddrsLAN
	c.NodeMeta = f.NodeMeta

	// if no bind address is given but ports are specified then we bail.
	// this only affects tests since in prod this gets merged with the
	// default config which always has a bind address.
	if f.BindAddr == nil && !reflect.DeepEqual(f.Ports, FilePorts{}) {
		return Config{}, fmt.Errorf("no bind address specified")
	}

	if f.BindAddr != nil {
		c.BindAddrs = []string{addrVal(f.BindAddr)}
	}

	if f.Ports.DNS != nil {
		c.DNSPort = intVal(f.Ports.DNS)
		for _, addr := range c.BindAddrs {
			c.DNSAddrsTCP = append(c.DNSAddrsTCP, joinHostPort(addr, c.DNSPort))
			c.DNSAddrsUDP = append(c.DNSAddrsUDP, joinHostPort(addr, c.DNSPort))
		}
	}

	return
}
