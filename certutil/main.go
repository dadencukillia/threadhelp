package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
)

func main() {
	if err := loadFromEnv(); err != nil {
		log.Fatal(err)
	}

	absPath, err := filepath.Abs(certDir)
	if err != nil {
		absPath = certDir
	}

	log.Printf(
		"Details:\n - Email: %s\n - Domain: %s\n - Webroot: %s\n - Cert save place: %s\n - Private key save place: %s\n",
		email, domain, webrootPath,
		filepath.Join(absPath, certFilename),
		filepath.Join(absPath, privateKeyFilename),
	)

	if err := os.MkdirAll(certDir, 0750); err != nil {
		log.Fatal(err)
	}

	client, err := newClient()
	if err != nil {
		log.Fatal(err)
	}

	notBefore, notAfter, succ := getCertTimes()
	if succ {
		now := time.Now()
		if now.Add(time.Hour).Before(notBefore.Add(autoRenewInterval)) && now.Add(time.Hour).Before(notAfter) {
			wait := notBefore.Add(autoRenewInterval).Sub(now.Add(time.Hour))
			log.Printf("Certificate creation in %.2f hours\n", wait.Hours())
			time.Sleep(wait)
		}
	}

	for {
		if err := obtainAndSaveCerts(client); err != nil {
			log.Printf("Error, retry in 3 minutes: %s\n", err)
			time.Sleep(3 * time.Minute)
		} else {
			log.Println("Certificate successfully saved!")
			log.Printf("Certificate renewal in %.2f hours\n", autoRenewInterval.Hours())
			time.Sleep(autoRenewInterval)
		}
	}
}

func getCertTimes() (notBefore time.Time, notAfter time.Time, success bool) {
	cont, err := os.ReadFile(fmt.Sprintf("%s/%s", certDir, certFilename))
	if err != nil {
		return time.Time{}, time.Time{}, false
	}

	block, _ := pem.Decode(cont)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}

	return cert.NotBefore, cert.NotAfter, true
}

func newClient() (*lego.Client, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	user := User{
		Email: email,
		key:   privateKey,
	}

	provider, err := webroot.NewHTTPProvider(webrootPath)
	if err != nil {
		return nil, err
	}

	config := lego.NewConfig(&user)

	config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	err = client.Challenge.SetHTTP01Provider(provider)
	if err != nil {
		return nil, err
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
	if err != nil {
		return nil, err
	}
	user.Registration = reg

	return client, nil
}

func obtainAndSaveCerts(client *lego.Client) error {
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificate, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fmt.Sprintf("%s/%s", certDir, certFilename), certificate.Certificate, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(fmt.Sprintf("%s/%s", certDir, privateKeyFilename), certificate.PrivateKey, 0600); err != nil {
		return err
	}

	return nil
}
