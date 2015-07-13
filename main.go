package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	hb "github.com/whyrusleeping/hellabot"
	shell "github.com/whyrusleeping/ipfs-shell"
)

var friends = []string{"whyrusleeping", "jbenet", "tperson", "krl", "kyledrake", "zignig"}

func isFriend(name string) bool {
	for _, n := range friends {
		if n == name {
			return true
		}
	}
	return false
}

type sayer interface {
	Say(string)
}

func Pin(s sayer, hash string) {
	s.Say(fmt.Sprintf("now pinning %s", hash))
	for i, sh := range shs {
		out, err := sh.Refs(hash, true)
		if err != nil {
			s.Say(fmt.Sprintf("[host %d] failed to grab refs for %s: %s", i, hash, err))
			return
		}

		// throw away results
		for range out {
		}

		err = sh.Pin(hash)
		if err != nil {
			s.Say(fmt.Sprintf("[host %d] failed to pin %s: %s", i, hash, err))
		}
	}
	s.Say(fmt.Sprintf("pin %s successful!", hash))
}

var EatEverything = &hb.Trigger{
	func(mes *hb.Message) bool {
		return true
	},
	func(irc *hb.IrcCon, mes *hb.Message) bool {
		return true
	},
}

var OmNomNom = &hb.Trigger{
	func(mes *hb.Message) bool {
		return mes.Content == "!botsnack"
	},
	func(irc *hb.IrcCon, mes *hb.Message) bool {
		irc.Channels[mes.To].Say("om nom nom")
		return true
	},
}

var authTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return true
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		if isFriend(mes.From) {
			// do not consume messages from authed users
			return false
		}
		return true
	},
}

var pinTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return isFriend(mes.From) && strings.HasPrefix(mes.Content, "!pin")
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		parts := strings.Split(mes.Content, " ")
		if len(parts) == 1 {
			con.Channels[mes.To].Say("what do you want me to pin?")
		} else {
			Pin(con.Channels[mes.To], parts[1])
		}
		return true
	},
}

var listTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return mes.Content == "!friends"
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		out := "my friends are: "
		for _, n := range friends {
			out += n + " "
		}
		con.Channels[mes.To].Say(out)
		return true
	},
}

var shs []*shell.Shell

func loadHosts() []string {
	fi, err := os.Open("hosts")
	if err != nil {
		fmt.Println("failed to open hosts file, defaulting to localhost:5001")
		return []string{"localhost:5001"}
	}

	var hosts []string
	scan := bufio.NewScanner(fi)
	for scan.Scan() {
		hosts = append(hosts, scan.Text())
	}
	return hosts
}

func main() {
	name := flag.String("name", "pinbot-test", "set pinbots name")
	flag.Parse()

	for _, h := range loadHosts() {
		shs = append(shs, shell.NewShell(h))
	}

	con, err := hb.NewIrcConnection("irc.freenode.net:6667", *name, false, true)
	if err != nil {
		panic(err)
	}

	con.AddTrigger(pinTrigger)
	con.AddTrigger(listTrigger)
	con.AddTrigger(OmNomNom)
	con.AddTrigger(EatEverything)
	con.Start()
	con.Join("#ipfs")

	for _ = range con.Incoming {
	}
}
