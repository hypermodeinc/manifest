# Hypermode Manifest

This repo hosts the [JSON schema](https://json-schema.org/) for `hypermode.json`.

[Hypermode](https://hypermode.com/home) | [Docs](https://docs.hypermode.com/manifest)

## Usage

In your `hypermode.json`, add this line to get autocomplete and error checking:

```json
{
  "$schema": "https://manifest.hypermode.com/hypermode.json"
}
```

## Extras

This repository also contains a small Go module, used internally at Hypermode
to ensure that reading from a `hypermode.json` manifest file is done consistently
across various projects.  It is not intended for public use.
