package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"golang.org/x/crypto/ed25519"
	"net/http"
	"sync"
	"time"
)

type GoTruePublicKey = ed25519.PublicKey
type GoTruePrivateKey = ed25519.PrivateKey

var _keyLock = sync.RWMutex{}
var _isKeysInitialized = false
var _publicKey GoTruePublicKey
var _privateKey GoTruePrivateKey
var _keyStorage jwkset.Storage
var _keyfunc keyfunc.Keyfunc

func initializeKeys(ctx context.Context) error {
	_keyLock.RLock()
	defer _keyLock.RUnlock()

	var err error
	_publicKey, _privateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	jwk, err := jwkset.NewJWKFromKey(_publicKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{
			ALG: jwkset.AlgEdDSA,
			USE: jwkset.UseSig,
		},
	})

	_keyStorage = jwkset.NewMemoryStorage()
	err = _keyStorage.KeyWrite(ctx, jwk)
	if err != nil {
		return err
	}

	_keyfunc, err = keyfunc.New(keyfunc.Options{
		Ctx:          ctx,
		Storage:      _keyStorage,
		UseWhitelist: []jwkset.USE{jwkset.UseSig},
	})
	if err != nil {
		return err
	}

	_isKeysInitialized = true
	return nil
}

func ensureInitialized(ctx context.Context) error {
	if !_isKeysInitialized {
		err := initializeKeys(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetPrivateKey(ctx context.Context) (GoTruePrivateKey, error) {
	if err := ensureInitialized(ctx); err != nil {
		return nil, err
	}

	return _privateKey, nil
}

func GetPublicKey(ctx context.Context) (GoTruePublicKey, error) {
	if err := ensureInitialized(ctx); err != nil {
		return nil, err
	}

	return _publicKey, nil
}

func GetKeyStorage(ctx context.Context) (jwkset.Storage, error) {
	if err := ensureInitialized(ctx); err != nil {
		return nil, err
	}

	return _keyStorage, nil
}

func GetKeyfunc(ctx context.Context) (keyfunc.Keyfunc, error) {
	if err := ensureInitialized(ctx); err != nil {
		return nil, err
	}

	return _keyfunc, nil
}

// Keys is the endpoint for acquiring JWKs
func (a *API) Keys(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	start := time.Now()

	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	storage, err := GetKeyStorage(ctx)
	if err != nil {
		return internalServerError("Failed to acquire JWKs storage")
	}

	marshaled, err := storage.Marshal(ctx)
	if err != nil {
		return internalServerError("Failed to marshal JWKs storage")
	}

	fmt.Printf("[api.Keys] Elapsed: %s\n", time.Since(start))

	return sendPrettyJSON(w, http.StatusOK, marshaled)
}
