// +build !windows

package alert

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const EnvAlertMailRecipient = "ALERT_MAIL_RECIPIENT"
const EnvAlertMailSender = "ALERT_MAIL_SENDER"
const EnvAlertMailTitlePrefix = "ALERT_MAIL_TITLE_PREFIX"

type mail struct {
	title   string
	message string
}

// Mail sends a mail to os.Getenv(alert.EnvAlertMailRecipient)
// from os.Getenv(alert.EnvAlertMailRecipient) using /usr/sbin/sendmail
// with a title prefix of os.Getenv(alert.EnvAlertMailTitlePrefix) + " " (if set).
func Mail(title string, format string, args ...interface{}) {
	recipient, sender := os.Getenv(EnvAlertMailRecipient), os.Getenv(EnvAlertMailSender)
	if strings.Count(recipient, "@") != 1 {
		log.Printf("alert.Mail(): cannot send mail: recipient from env var %s is malformed.", EnvAlertMailRecipient)
		return
	}
	if strings.Count(sender, "@") != 1 {
		log.Printf("alert.Mail(): cannot send mail: sender from env var %s is malformed.", EnvAlertMailSender)
		return
	}
	fullTitle := strings.TrimSpace(os.Getenv(EnvAlertMailTitlePrefix) + " " + title)
	err := sendMail(&mail{title: fullTitle, message: fmt.Sprintf(format, args...)}, recipient, sender)
	if err != nil {
		log.Printf("alert.Mail(): cannot send mail: %v", err)
	}
}

func sendMail(m *mail, recipient, sender string) (retErr error) {
	sendmail := exec.Command("/usr/sbin/sendmail", "-f", sender, recipient)
	stdin, err := sendmail.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not get StdinPipe: %w", err)
	}
	defer stdin.Close()
	stdout, err := sendmail.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get StdoutPipe: %w", err)
	}
	defer stdout.Close()
	err = sendmail.Start()
	if err != nil {
		return fmt.Errorf("could not start /usr/sbin/sendmail: %w", err)
	}
	defer func() {
		output, stdoutErr := ioutil.ReadAll(stdout)
		err = sendmail.Wait()
		if retErr != nil {
			return
		}
		if err != nil {
			retErr = fmt.Errorf("could not wait for process /usr/sbin/sendmail to finish: %w", err)
			return
		}
		if !sendmail.ProcessState.Success() {
			if stdoutErr == nil {
				retErr = fmt.Errorf("sendmail failed with exit code %d: %v", sendmail.ProcessState.ExitCode(), string(output))
			} else {
				retErr = fmt.Errorf("sendmail failed with exit code %d; also couldn't read its stdout: %v", sendmail.ProcessState.ExitCode(), stdoutErr)
			}
			return
		}
	}()
	if _, err = stdin.Write([]byte("Subject: " + m.title + "\n\n" + m.message)); err != nil {
		return fmt.Errorf("could not write in stdin of /usr/sbin/sendmail: %w", err)
	}
	if err = stdin.Close(); err != nil {
		return fmt.Errorf("could not close stdin of /usr/sbin/sendmail: %w", err)
	}
	return nil
}
