# Sopsistry

A command-line tool for managing team-based secret encryption, built on top of
[SOPS](https://github.com/mozilla/sops) and [age](https://github.com/FiloSottile/age).


## Rationale

SOPS provides robust cryptographic capabilities for secret management but lacks support for team-oriented workflows.
With SOPS, you have basically static configuration without connection between keys and member identities and roles:

```yaml
# .sops.yaml
creation_rules:
  - key_groups:
    - age:
      - age1abc123...  # Identity unknown
      - age1def456...  # Employment status unclear
      - age1ghi789...  # Access validity uncertain
```

Sopsistry introduces explicit team member mapping:

```yaml
# sopsistry.yaml
members:
  - id: alice
    age_key: age1abc123...
  - id: bob
    age_key: age1def456...
  - id: charlie
    age_key: age1ghi789...
...
scopes:
  - label: production
    members:
      - alice
      - charlie
    files:
      - secrets/prod.yaml
  - label: development
    members:
      - bob
    files:
      - secrets/dev.yaml
default_scopes: [development]
```

Sopsistry is intended to be for SOPS kind of like what docker-compose is to docker.
You don't need to know how to use SOPS, but Sopsistry can coexist with it -- it just makes
it easier for most common uses cases within a team.


## Prerequisites

- [SOPS](https://github.com/mozilla/sops)
- [age](https://github.com/FiloSottile/age) for secret encryption/decryption
- Project managed with Git

## Installation

```bash
go install github.com/edvardm/sopsistry@latest
```

## Quick Start

```bash
# initialize sopsistry config file. Adds current user automatically
sistry init

# Add team member
sistry member add alice age1abc123...

# Encrypt whoel file -- no need to specify keys
sistry encrypt secrets.yaml

# Encrypt matching values only
sistry encrypt --help --iregex '(password|token)' prod.env

# Decrypt file to stdout, works regardless of who encrypted the file, as long
# as user is member of the team
sistry decrypt prod.envs
```

## TODO
- support at least KMS to decrypt secrets in typical cloud deploy environments
- super easy secret rotation
- TUI?
- optional scopes, so that by default secrets are shared for dev-env only by default,
  staging/prod access requires separate access etc
- extend `check` to ensure configuration is valid, e.g. all members have age keys, user's
  age key is in the config, key is not expired etc


## FAQ

- Support GnuPG as backend? or ssh keys?
  - It is better to have project-specific keys, which is also why Sopsistry
    stores age keys in hidden project directory, user-readable only. So probably not.
    Also, age is super simply to install and use.
- How is it possible that I can decrypt file even if somebody else encrypted it?
  - When encrypting a secrets, SOPS creates random symmetric key, which is then
    encrypted with all team members' age keys, allowing any team member to decrypt that random key.
    Then any secrets are just decrypted with that now decrypted symmetric key.
- Our team uses LastPass/1Password/Bitwarden, could we use that as backend?
  - Even though most of these allow programmatic access, accidentally leaking
    such secret would be much more dangerous than leaking ephemeral, software
    project specific key which is expected to change frequently.


    It is not intended to replace password managers, but rather to complement them.


## License

MIT License - see [LICENSE](LICENSE) for details.