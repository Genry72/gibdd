package utils

import (
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)
//CopyFileToHost Копируем локальный файл на удаленную площадку
func CopyFileToHost(srcPath, dstPath, username, sshKeyPath, hostname, test string) (err error) {
	var port string
	var config *ssh.ClientConfig
	if test == "false" { //Если катим удаленно
		port = "22"
		// SSH client config
		config = &ssh.ClientConfig{
			Timeout:         time.Second, //ssh connection time out time is one second, if SSH validation error returns in one second
			User:            username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // This is OK, but not safe enough.
			//HostKeyCallback: hostKeyCallBackFunc(h.Host),
		}
		config.Auth = []ssh.AuthMethod{PublicKeyAuthFunc(sshKeyPath)}
	} else {//катим на локальный тестовый образ
		port = "22"
		hostname = "10.10.50.15"
		config = &ssh.ClientConfig{
			User: "valentin",
			Auth: []ssh.AuthMethod{
				ssh.Password("Phetore22"),
			},
			// Non-production only
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}
	client, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		return
	}
	defer client.Close()

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		return
	}
	defer dstFile.Close()

	// write to file
	if _, err = dstFile.ReadFrom(srcFile); err != nil {
		return
	}
	return
}
