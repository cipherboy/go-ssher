package main

import (
  "fmt"
  "golang.org/x/crypto/ssh"
  "sync"
  "bufio"
  "os"
  "strings"
)

func main()  {
  if len(os.Args) != 4 {
    fmt.Println("Usage: go-ssher /path/to/hosts.txt /path/to/users.txt /path/to/passwords.txt")
    return
  }

  var hosts []string
  var users []string
  var passwords []string

  hosts_f, err := os.Open(os.Args[1])
  if err != nil {
    fmt.Println(err)
    return
  }
  defer hosts_f.Close()

  users_f, err := os.Open(os.Args[2])
  if err != nil {
    fmt.Println(err)
    return
  }
  defer users_f.Close()

  passwords_f, err := os.Open(os.Args[3])
  if err != nil {
    fmt.Println(err)
    return
  }
  defer passwords_f.Close()

  hosts_s := bufio.NewScanner(hosts_f)
  for hosts_s.Scan() {
    hosts = append(hosts, hosts_s.Text())
  }
  if hosts_s.Err() != nil {
    fmt.Println(hosts_s.Err())
    return
  }

  users_s := bufio.NewScanner(users_f)
  for users_s.Scan() {
    users = append(users, users_s.Text())
  }
  if users_s.Err() != nil {
    fmt.Println(users_s.Err())
    return
  }

  passwords_s := bufio.NewScanner(passwords_f)
  for passwords_s.Scan() {
    passwords = append(passwords, passwords_s.Text())
  }
  if passwords_s.Err() != nil {
    fmt.Println(passwords_s.Err())
    return
  }

  var wg sync.WaitGroup
  for u := range users {
    user := users[u]
    for p := range passwords {
      password := passwords[p]
      fmt.Println("User:", user, "Password:", password)
      sshConfig := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
        	ssh.Password(password),
        },
      }


      for h := range hosts {
        wg.Add(1)
        var system string = hosts[h]
        if !strings.Contains(system, ":") {
          system += ":22"
        }
        
        go func(host string, Username string, Password string) {
          connection, err := ssh.Dial("tcp", host, sshConfig)
          if err != nil {
            fmt.Println(host, "- no")
            wg.Done()
            return
          }

          session, err := connection.NewSession()
          if err != nil {
            fmt.Println(host, "- no")
            wg.Done()
            return
          }

          modes := ssh.TerminalModes{
            ssh.ECHO:          0,     // disable echoing
            ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
            ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
          }

          if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
            session.Close()
            fmt.Println(host, "- no")
            wg.Done()
            return
          }

          err = session.Run("hostname")
          if err != nil {
            wg.Done()
            return
          }

          fmt.Println(host, "- yes, u:", Username, ", p:", Password)
          wg.Done()
        }(system, user, password)
      }

    }
  }
  wg.Wait()
}
