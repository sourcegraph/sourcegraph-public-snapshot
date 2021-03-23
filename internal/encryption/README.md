# Encryption

This package provides tools to encrypt & decrypt data via the encryption.Key interface. This interface is built to wrap any encryption backend, such as cloud provider APIs, stdlib encryption libraries, or testing stubs. 

This package was originally designed in [RFC 310](https://docs.google.com/document/d/1ZlQzlTRtrQbx3yi2cqmSjyq3ddcp2eKnhqekLOvzm_w/edit#)

### How to use this package
Keys should be passed in/set during initialisation, ideally from `main()`. Accessing `keyring.Default` is an antipattern, and only provided for cases where injection is particularly difficult. You should also only pass the individual key(s) that you need, rather than the whole ring.

Data should be kept encrypted for as long as possible. Right now our implementations are decrypting data inside the `database` package, which is an antipattern. We made this choice in order to make progress as the code did not lend itself easily to moving the decryption out of the store. Ideally you keep the data encrypted until it is passed to whatever needs the zero visibility data.

### Zero Visibility Data (`encryption.Secret`)
The plaintext returned by the `Key.Decrypt()` method is considered 'zero visibility data'. This means that no human should ever be able to see this data, and if someone does it should be considered compromised, and be replaced. In order to make accidental disclosure more difficult the encryption package returns data in an `encryption.Secret` wrapper type. This type wraps a value in a struct with an unexported field, implementing the `Stringer` & `json.Marshaler` interfaces & redacting the data. The only method that returns the plaintext is `Secret.Secret()`, this means our handling of secrets is more auditable, and reduces the chances of accidentally leaking the value in logs.

### Keyring
The `encryption/keyring` package provides a way to configure encryption keys & retrieve them in a typesafe manner, it parses site config and sets the keys in a `keyring.Ring` struct, so users can either access the `keyring.Default` or inject the ring, and access specific keys safely, rather than needing to spread around the concern of correctly configuring a key.

### Composition & extension
The `encryption.Key` interface was built to be simple, and intended to be extended through composition & embedding. For example we plan to enable key migrations using a Key implementation that wraps two other Keys, decrypting with one & encrypting with the other. You could also create an encryption.Key wrapper that implements it's own versioning system, encrypting with a 'primary' Key, but being able to decrypt data with the previous keys.

### Implementations

* Cloud KMS
* Mounted Key
* No Op
