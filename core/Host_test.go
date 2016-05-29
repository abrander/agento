package core

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/transports/local"
	"github.com/abrander/agento/plugins/transports/ssh"
	"github.com/abrander/agento/userdb"
)

type Config struct {
	Hosts map[string]toml.Primitive `toml:"host"`
}

func TestHostDecodeTOML(t *testing.T) {
	cases := map[string]*Host{
		`[host.testhost]
        transport = "localtransport"
        name = "testhost"
        `: &Host{
			AccountID: userdb.God.GetAccountId(),
			Name:      "testhost",
			Transport: localtransport.NewLocalTransport().(plugins.Transport),
		},
		`[host.testhost]
        transport = "sshtransport"
        name = "testhost2"
        host = "dev.google.com"
        `: &Host{
			AccountID: userdb.God.GetAccountId(),
			Name:      "testhost2",
			Transport: &ssh.SshTransport{
				Ssh: ssh.Ssh{
					Host: "dev.google.com",
					Port: 22,
				},
			},
		},
		`[host.testhost]
        transport = "sshtransport"
        host = "127.0.0.1"
        port = 200
        `: &Host{
			AccountID: userdb.God.GetAccountId(),
			Transport: &ssh.SshTransport{
				Ssh: ssh.Ssh{
					Host: "127.0.0.1",
					Port: 200,
				},
			},
		},
		`[host.testhost]
        transport = "nonexisting"
        name = "testhost4"
        `: nil,
	}

	for conf, correct := range cases {
		c := Config{}
		_, err := toml.Decode(conf, &c)

		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}

		for _, prim := range c.Hosts {
			host := Host{}
			err = host.DecodeTOML(prim)
			if err != nil && correct != nil {
				t.Fatalf("DecodeTOML error: %s", err.Error())
			}

			if correct == nil {
				if err == nil {
					t.Errorf("DecodeTOML didn't catch error")
				}
				break
			}

			if len(host.ID) != 20 {
				t.Error("Failed to generate ID")
			}

			// Because of id randomization, we have to override this.
			host.ID = ""

			tStr := fmt.Sprintf("%s", host.Transport)
			cStr := fmt.Sprintf("%s", correct.Transport)

			if tStr != cStr {
				t.Errorf("Host transport %+v doesn't match correct transport %+v", host.Transport, correct.Transport)
			}

			host.Transport = nil
			correct.Transport = nil

			if host != *correct {
				t.Errorf("Host %+v doesn't match correct %+v", host, correct)
			}
		}
	}
}

func TestHostDecodeJSON(t *testing.T) {
	cases := map[string]*Host{
		`{
		  "_id": "8764786ab76dbc76e",
		  "name": "testhost",
		  "transport": "localtransport",
		  "accountId": "hejhejsa"
		}`: &Host{
			ID:        "8764786ab76dbc76e",
			AccountID: "hejhejsa",
			Name:      "testhost",
			Transport: localtransport.NewLocalTransport().(plugins.Transport),
		},
		`{
		  "_id": "",
		  "name": "testhost2",
		  "transport": "sshtransport",
		  "host": "dev.google.com"
	    }`: &Host{
			Name: "testhost2",
			Transport: &ssh.SshTransport{
				Ssh: ssh.Ssh{
					Host: "dev.google.com",
					Port: 22,
				},
			},
		},
		`{
		  "transport": "sshtransport",
		  "host": "127.0.0.1",
		  "port": 200
		}`: &Host{
			Transport: &ssh.SshTransport{
				Ssh: ssh.Ssh{
					Host: "127.0.0.1",
					Port: 200,
				},
			},
		},
		`{
		  "transport": "nonexisting",
	    }`: nil,
		`invalid json`: nil,
	}

	for data, correct := range cases {
		host := Host{}
		err := json.Unmarshal([]byte(data), &host)
		if err != nil && correct != nil {
			t.Fatalf("json.Unmarshal error: %s", err.Error())
		}

		if correct == nil {
			if err == nil {
				t.Errorf("DecodeTOML didn't catch error")
			}
			break
		}

		tStr := fmt.Sprintf("%s", host.Transport)
		cStr := fmt.Sprintf("%s", correct.Transport)

		if tStr != cStr {
			t.Errorf("Host transport %+v doesn't match correct transport %+v", host.Transport, correct.Transport)
		}

		host.Transport = nil
		correct.Transport = nil

		if host != *correct {
			t.Errorf("Host %+v doesn't match correct %+v", host, correct)
		}
	}
}
