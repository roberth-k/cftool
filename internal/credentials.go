package internal

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type cachedCredentials struct {
	Credential credentials.Value
	Expiration time.Time
	Profile    string
}

func (c *cachedCredentials) IsExpired() bool {
	return c.Expiration.Before(time.Now())
}

func getCacheDir() string {
	var homedir string

	if runtime.GOOS == "windows" {
		homedir = os.Getenv("APPDATA")
	} else {
		homedir = filepath.Join(os.Getenv("HOME"), ".cache")
	}

	dirpath := filepath.Join(homedir, "cftool", "credentials")

	_, err := os.Stat(dirpath)

	if os.IsNotExist(err) {
		_ = os.MkdirAll(dirpath, os.ModeDir|0700)
	}

	return dirpath
}

func WrapCredentialsWithCache(
	profile string,
	creds *credentials.Credentials,
) (*credentials.Credentials, error) {
	provider := NewCachedCredentialProvider(profile, creds)
	return credentials.NewCredentials(provider), nil
}

func NewCachedCredentialProvider(
	profile string,
	creds *credentials.Credentials,
) credentials.Provider {
	if profile == "" {
		profile = os.Getenv("AWS_PROFILE")
	}

	hash := md5.New()
	_, _ = io.WriteString(hash, profile)
	digest := hex.EncodeToString(hash.Sum(nil))
	credpath := filepath.Join(getCacheDir(), digest+".json")

	cp := &cachedCredentialProvider{creds, cachedCredentials{}, credpath, profile}
	cp.read()
	return cp
}

var _ credentials.Provider = (*cachedCredentialProvider)(nil)

type cachedCredentialProvider struct {
	inner    *credentials.Credentials
	outer    cachedCredentials
	credpath string
	profile  string
}

func (my *cachedCredentialProvider) read() {
	_, err := os.Stat(my.credpath)

	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Panicf("oops: stat %s (%v)", my.credpath, err)
	}

	data, err := ioutil.ReadFile(my.credpath)

	if err != nil {
		log.Panicf("oops: read %s", my.credpath)
	}

	var creds cachedCredentials
	err = json.Unmarshal(data, &creds)

	if err != nil {
		log.Panicf("oops: unmarshal %s", my.credpath)
	}

	if creds.IsExpired() {
		return
	}

	my.outer = creds
}

func (my *cachedCredentialProvider) write(v credentials.Value, exp time.Time) {
	cc := cachedCredentials{v, exp, my.profile}
	data, err := json.Marshal(&cc)
	if err != nil {
		log.Panicf("oops: write credentials (%v)", err)
	}

	err = ioutil.WriteFile(my.credpath, data, 0600)

	if err != nil {
		log.Panicf("oops: write %s (%v)", my.credpath, err)
	}

	my.outer = cc
}

func (my *cachedCredentialProvider) Retrieve() (credentials.Value, error) {
	if !my.outer.IsExpired() {
		return my.outer.Credential, nil
	} else {
		v, err := my.inner.Get()

		if err == nil {
			exp, _ := my.inner.ExpiresAt()
			my.write(v, exp)
		}

		return v, err
	}
}

func (my *cachedCredentialProvider) IsExpired() bool {
	return my.outer.IsExpired() && my.inner.IsExpired()
}
