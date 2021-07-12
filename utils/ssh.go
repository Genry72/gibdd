package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	// Uncomment to store output in variable
	//"bytes"
)

// username := "valsemenov86"
// sshKeyPath := "./id_rsa"
// hostname := "104.196.182.179"
//sshExec Выполнение коменд на удаленном хосте
//sshExec Выполнение коменд на удаленном хосте
func sshExec(hostname, sshKeyPath, username string, commands []string) (err error) {
	port := "22"
	// SSH client config
	config := &ssh.ClientConfig{
		Timeout:         time.Second, //ssh connection time out time is one second, if SSH validation error returns in one second
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // This is OK, but not safe enough.
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}
	config.Auth = []ssh.AuthMethod{PublicKeyAuthFunc(sshKeyPath)}
	// Connect to host
	client, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create sesssion
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		return err
	}

	// Uncomment to store output in variable
	var b bytes.Buffer
	//sess.Stdout = &b
	sess.Stderr = &b

	// Enable system stdout
	// Comment these if you uncomment to store in variable
	sess.Stdout = os.Stdout
	// sess.Stderr = os.Stderr

	// Start remote shell
	err = sess.Shell()
	if err != nil {
		return err
	}

	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			return err
		}
	}

	// Wait for sess to finish
	err = sess.Wait()
	if err != nil {
		return
	}

	if b.String() != "" {
		err = fmt.Errorf("ошибка выполнения команд:\n%v", b.String())
		return
	}
	return
}

//PublicKeyAuthFunc возвращает публичный ключь
func PublicKeyAuthFunc(kPath string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(kPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}
