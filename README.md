## Alert

Send alert mails, configured via environment variables. UNIX only. Requires `/usr/sbin/sendmail`.

```sh
export ALERT_MAIL_RECIPIENT=alert@example.com
export ALERT_MAIL_SENDER=service@example.com
export ALERT_MAIL_TITLE_PREFIX=[PRODUCTION]
```

```golang
import "github.com/setlog/alert"

func main() {
    alert.Mail("Import failed!", "Failed to parse line %d", 42) // fmt.Sprintf-style formatting.
}
```
