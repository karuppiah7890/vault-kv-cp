# vault-kv-cp

## Building

```bash
go build -v
```

## Usage

```bash
$ ./vault-kv-cp
usage: vault-kv-cp <source-kv-mount-path> <destination-kv-mount-path>
```

## Demo

Source Vault, secured with HTTPS and a big token. It has some secrets - with one secret having many versions (3 versions)

```bash

$ vault secrets list
Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_b8f3ed6a    per-token private secret storage
identity/     identity     identity_e3c8f173     identity store
secret/       kv           kv_86bfa42e           Generic secret storage
sys/          system       system_9a2cc83b       system endpoints used for control, policy and debugging

$ vault kv list
Not enough arguments (expected 1, got 0)

$ vault kv list secret
Keys
----
demosecret/
something/

$ vault kv list secret/demosecret
Keys
----
aws

$ vault kv list secret/demosecret/aws
No value found at secret/metadata/demosecret/aws

$ vault kv get secret/demosecret/aws
======= Secret Path =======
secret/data/demosecret/aws

======= Metadata =======
Key                Value
---                -----
created_time       2024-03-21T15:22:17.814449057Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            3

============ Data ============
Key                      Value
---                      -----
AWS_SECRET_ACCESS_KEY    s3cr3t
PASSWORD                 somePasswordHere
TOKEN                    blah

$ vault kv list secret/
Keys
----
demosecret/
something/

$ vault kv list secret/something
Keys
----
else/

$ vault kv list secret/something/else
Keys
----
wow

$ vault kv list secret/something/else/wow
No value found at secret/metadata/something/else/wow

$ vault kv get secret/something/else/wow
========= Secret Path =========
secret/data/something/else/wow

======= Metadata =======
Key                Value
---                -----
created_time       2024-03-21T15:22:02.034111335Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

============ Data ============
Key                      Value
---                      -----
AWS_SECRET_ACCESS_KEY    lol
PASSWORD                 somePasswordHere
TOKEN                    SECRETIVE_API_TOKEN_HERE
```

Destination Vault, local dev Vault server, with no HTTPS, with token as `root`

```bash
$ export VAULT_TOKEN="root"

$ export VAULT_ADDR='http://127.0.0.1:8300'

$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         1.15.4
Build Date      2023-12-04T17:45:28Z
Storage Type    inmem
Cluster Name    vault-cluster-7bc4c9ee
Cluster ID      fee68bf6-8eff-4cfd-bdab-abd6ef1b3282
HA Enabled      false

$ vault secrets list
Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_e433e0c2    per-token private secret storage
identity/     identity     identity_0e7a4812     identity store
secret/       kv           kv_8e374477           key/value secret storage
sys/          system       system_271e70ee       system endpoints used for control, policy and debugging

$ vault kv list
Not enough arguments (expected 1, got 0)

$ vault kv list secret
No value found at secret/metadata
```

Now, let's copy from source to destination :) The latest version of the secrets gets copied from the source to the destination. Any secrets in the destination is overwritten

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'

$ export SOURCE_VAULT_CACERT=$PWD/vault-ca.crt

$ export SOURCE_VAULT_TOKEN="big-token-here"

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'

$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-kv-cp secret secret
&{RequestID:0c9f6194-fcfb-c332-8f2e-fe70cd80afc1 LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[demosecret/ something/]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

&{RequestID:61e18f2b-b471-6c34-28ab-069ecf5cc4a3 LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[aws]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

copying secret at `demosecret/aws` in source to `secret` in destination

&{RequestID:5e67e48c-e8ac-c733-b3af-462d8dcd2bb0 LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[else/]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

&{RequestID:66144652-3f84-876c-0fc2-bad5dcac6ba5 LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[wow]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

copying secret at `something/else/wow` in source to `secret` in destination


$ vault kv list secret
Keys
----
demosecret/
something/

$ vault kv list secret/demosecret
Keys
----
aws

$ vault kv list secret/demosecret/aws
No value found at secret/metadata/demosecret/aws

$ vault kv get secret/demosecret/aws
======= Secret Path =======
secret/data/demosecret/aws

======= Metadata =======
Key                Value
---                -----
created_time       2024-03-21T15:34:02.203551Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

============ Data ============
Key                      Value
---                      -----
AWS_SECRET_ACCESS_KEY    s3cr3t
PASSWORD                 somePasswordHere
TOKEN                    blah

$ vault kv list secret/
Keys
----
demosecret/
something/

$ vault kv list secret/something
Keys
----
else/

$ vault kv list secret/something/else
Keys
----
wow

$ vault kv list secret/something/else/wow
No value found at secret/metadata/something/else/wow

$ vault kv get secret/something/else/wow
========= Secret Path =========
secret/data/something/else/wow

======= Metadata =======
Key                Value
---                -----
created_time       2024-03-21T15:34:02.800351Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

============ Data ============
Key                      Value
---                      -----
AWS_SECRET_ACCESS_KEY    lol
PASSWORD                 somePasswordHere
TOKEN                    SECRETIVE_API_TOKEN_HERE

$ ### Now let's overwrite the secrets in the destination by doing the copy again :)

$ ./vault-kv-cp secret secret
&{RequestID:a5915878-ca1e-2308-3dbe-f589dfaa3e28 LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[demosecret/ something/]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

&{RequestID:1fa5911a-bfc2-6d9f-3b1d-df0a9061467d LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[aws]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

copying secret at `demosecret/aws` in source to `secret` in destination

&{RequestID:66ab418d-e953-f72f-021b-d1c238ff72af LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[else/]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

&{RequestID:2e20953a-a0b2-cea6-4745-3285c3f0a938 LeaseID: LeaseDuration:0 Renewable:false Data:map[keys:[wow]] Warnings:[] Auth:<nil> WrapInfo:<nil> MountType:}

copying secret at `something/else/wow` in source to `secret` in destination


$ vault kv get secret/something/else/wow
========= Secret Path =========
secret/data/something/else/wow

======= Metadata =======
Key                Value
---                -----
created_time       2024-03-21T15:35:39.809718Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            2

============ Data ============
Key                      Value
---                      -----
AWS_SECRET_ACCESS_KEY    lol
PASSWORD                 somePasswordHere
TOKEN                    SECRETIVE_API_TOKEN_HERE
```

## Future Ideas

- Enable KV v2 Secrets Engine in the destination Vault in the given destination mount path if it doesn't exist
- Copy KV v2 Secrets Engine configuration from source to destination
