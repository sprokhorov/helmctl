# helmctl

helmctl is a wrapper of helm, helm-diff and sops. It uses yaml formated files to describe and re-use helm releases.

## Installing

```shell
brew install go && \
git clone https://github.com/sprokhorov/helmctl.git && \
cd helmctl && go install github.com/sprokhorov/helmctl
```

## Using the helmctl

### Create helmctl.yaml

First of all we need helmctl.yaml to describe our releases. In following example we deploy `telegraf` on `development` environment. Please check [example-helmctl.yaml](example-helmctl.yaml) to see all possible options of release describing.
```yaml
version: v1
spec:
  releases:
    - name: telegraf
      chart: stable/telegraf
      version: 1.6.1
      values:
        - name: image.tag
          value: latest
  installs:
    environments:
      development:
        - telegraf
```
Syntax features:
* You can use environment variables with custom YAML tag `!env`
* You can include yaml blocks from files  with custom YAML tag `!include`
```yaml
version: v1
spec:
  releases:
  # Include values from file
    - !include releases/example/example-release-2.yaml
  # You can use `<<:` yaml merge operation to merge data from file with defined values
    - <<: !include releases/example/example-release-2.yaml
      name: example-release-2
```
* You can use override/append release params for each environemnt or project in installs section.
  Arrays will be appended, string values will be overrided:
```yaml
version: v1
spec:
  releases:
    - name: telegraf
      chart: stable/telegraf
      version: 1.6.1
      values:
        - name: image.tag
          value: latest
  installs:
    environments:
      development:
        - name: telegraf
          values:
          - name: image.pullPolicy
            value: always
          version: 1.6.2-NEW_VERSION
```
This will produce next release for development for `telegraf`:
```yaml
name: telegraf
chart: stable/telegraf
version: 1.6.2-NEW_VERSION
values:
  - name: image.tag
    value: latest
  - name: image.pullPolicy
    value: always
```

### Install releases

We can install one release to environment:
```shell
helmctl --environment development install telegraf
```
or all releases at the same time:
```shell
helmctl --environment development install all
```
