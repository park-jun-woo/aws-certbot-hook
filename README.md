# aws-certbot-hook – AWS Route53 Automatic DNS-01 Certbot Hook

[![MIT License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**aws-certbot-hook** is a lightweight Go binary to automate Let's Encrypt certbot DNS-01 validation on AWS Route53 domains.

* Automatically detects and updates the correct Route53 HostedZone for any domain, including public suffixes (e.g. `.com`, `.co.kr`).
* Works as a certbot --manual-auth-hook and --manual-cleanup-hook.
* Fast, dependency-free, portable, and production-ready.

---

## ✨ Features

* **Go single binary** (no runtime dependencies)
* **Automatic HostedZone detection** (publicsuffix-aware)
* **Works with AWS credentials via environment, shared config, or IAM Role**
* **Handles any domain structure: foo.bar.domain.com, example.co.kr, etc.**
* **Crontab + certbot renew** automation friendly

---

## 🚀 Installation & Build

```bash
git clone https://github.com/park-jun-woo/aws-certbot-hook.git
cd aws-certbot-hook

go mod tidy
make all    # build and install to /usr/local/bin
```

You should now have `/usr/local/bin/aws-certbot-hook` and `/usr/local/bin/acertbot` (wrapper script).

---

## 🔑 Requirements

* AWS CLI must be working (`aws sts get-caller-identity` should succeed)
* IAM Policy with at least:

  * `route53:ListHostedZones`
  * `route53:ChangeResourceRecordSets`
* For EC2: instance IAM Role is recommended

---

## ⚡️ Usage

Basic example for single or multi-domain:

```bash
acertbot -d cms-ec2.domain.com --email you@domain.com
```

Manual invocation:

```bash
certbot certonly \
  --manual \
  --preferred-challenges dns \
  --manual-auth-hook "/usr/local/bin/aws-certbot-hook --hook=auth" \
  --manual-cleanup-hook "/usr/local/bin/aws-certbot-hook --hook=cleanup" \
  --non-interactive --agree-tos --manual-public-ip-logging-ok \
  -d cms-ec2.domain.com
```

> **Tip:** Multiple `-d` domains (including wildcard) are fully supported.

---

## 🛠 Options

* `--hook=auth`    : Insert TXT record (for DNS-01 validation)
* `--hook=cleanup` : Remove TXT record (after validation)
* `--sleep=20`     : Wait N seconds for DNS propagation (default: 20)

---

## 📝 Example: crontab auto-renewal

```cron
0 3 * * * certbot renew --manual-auth-hook "aws-certbot-hook --hook=auth" --manual-cleanup-hook "aws-certbot-hook --hook=cleanup" --deploy-hook "systemctl reload nginx"
```

---

## 🧩 Makefile Targets

* `make build`   : Build the aws-certbot-hook binary
* `make install` : Install to `/usr/local/bin`
* `make setup`   : Update system root CAs (optional, for older distros)
* `make clean`   : Clean build artifacts

---

## 🔒 LICENSE

MIT License (c) 2025 PARK JUN WOO

---

## 🙋‍♂️ Contact / Contributing

* Issues, feature requests, PRs: [github.com/park-jun-woo/aws-certbot-hook/issues](https://github.com/park-jun-woo/aws-certbot-hook/issues)
* Bug reports or support: contact via GitHub Issues or parkjunwoo.com

---

> **Production-grade AWS Route53 DNS-01 certbot hook. Ready for SaaS, multi-domain, and global deployment.**
