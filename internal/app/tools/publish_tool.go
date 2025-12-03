package tools

type PublishTool struct{}

func (t *PublishTool) Publish(changeId, revision string, payload map[string]interface{}) (bool, error) {
    return true, nil
}

