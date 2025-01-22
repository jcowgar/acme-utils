# Acme Ollama Design Notes

## Editing Interface

The actual text file representing an interaction may look like:

```
# Learn about the sky

## You

What makes the sky blue?

### Response

The sky is blue because it is the color of the sky.

## You

That really does not answer my question

### Response

I'm sorry.
```

At some future point, we may want to provide some parameters as front matter to control Ollama. For
example, maybe specifying the model or temperature.

## Integrating with Acme

We can watch the log. When a put occurs on a file whos name contains +ollama, we read the file in.

After reading we should parse the content and get the last `## You` block and send the content to Ollama.

Ollama will then respond and we should update the buffer with it's response, after adding a `### Response`
header.

## Interacting with Ollama

Ollama provides a http api: [ollama api](https://github.com/ollama/ollama/blob/main/docs/api.md)