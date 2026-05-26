# Hibot CLI Examples

End-to-end recipes that use the `hibot` binary. Each section assumes you have
configured credentials via flags, environment variables, or
`$HOME/.hibot/config.yaml` (see the parent [`README.md`](../README.md)).

## 0. One-time setup

```
hibot --endpoint=https://open.volcengineapi.com \
      --ak=$HIBOT_AK --sk=$HIBOT_SK \
      --workspace-id=$HIBOT_WORKSPACE_ID --region=cn-beijing \
      config init
```

## 1. Create an Agent and chat

```
# pick a base model
MODEL_ID=$(hibot models providers list-models -o json | jq -r '.Items[0].ID')

# create an Agent
AGENT_ID=$(hibot agents create \
    --name=demo \
    --model-id=$MODEL_ID \
    --system "You are a friendly assistant." \
    -o json | jq -r .ID)

# create a Session and stream a chat
SESSION_ID=$(hibot sessions create --agent-id=$AGENT_ID -o json | jq -r .ID)
echo "Hello, who are you?" | hibot chat $SESSION_ID --stream
```

## 2. Upload a Skill bundle

```
hibot skills upload \
    --name=my-skill \
    --version=1.0.0 \
    --file=./skill.zip \
    --description="Custom Python skill"

hibot skills list
hibot skills versions --skill-id=<skill-id>
```

## 3. Register an MCP server

```
hibot mcps create \
    --name=weather \
    --transport=streamable-http \
    --endpoint=https://example.com/mcp \
    --header "Authorization=Bearer xyz" \
    --header "X-Tenant=acme"

# Test connection without persisting:
hibot mcps test --endpoint=https://example.com/mcp --header "Authorization=Bearer xyz"
```

## 4. Resources and Directories

```
DIR_ID=$(hibot resources directories create --name=docs -o json | jq -r .ID)

hibot resources create \
    --name=manual.pdf \
    --type=document_collection \
    --directory-id=$DIR_ID \
    --file=./manual.pdf

hibot resources list --directory-id=$DIR_ID
```

## 5. Prompt templates

```
hibot prompts create --name=summarizer --content @prompts/summarizer.md
hibot prompts list
hibot prompts update <prompt-id> --content @prompts/summarizer-v2.md
```

## 6. Environments

```
hibot environments create \
    --name=hermes-default \
    --image-type=hermes \
    --cpu=1 --memory=2Gi \
    --env-vars @env.json
```

## 7. Output formats

```
hibot agents list                  # default table
hibot agents list -o json | jq .
hibot agents list -o yaml
```

## 8. Verbose streaming

```
hibot chat $SESSION_ID --stream -v
# prints intermediate [event:tool_start] / [event:tool_complete] markers
```
