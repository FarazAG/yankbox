package clipboard

import (
	"errors"
	"fmt"

	"github.com/deploymenttheory/go-bindings-winrt/bindings/winrt/applicationmodel/datatransfer"
)

var ErrHistoryDisabled = errors.New("Windows Clipboard History is disabled")

type Item struct {
	ID   string
	Text string
}

func History() ([]Item, error) {
	clipboard, err := datatransfer.ClipboardStatics2()
	if err != nil {
		return nil, err
	}

	enabled, err := clipboard.IsHistoryEnabled()
	if err != nil {
		return nil, err
	}

	if !enabled {
		return nil, ErrHistoryDisabled
	}

	operation, err := clipboard.GetHistoryItemsAsync()
	if err != nil {
		return nil, err
	}

	result, err := operation.Await()
	if err != nil {
		return nil, err
	}

	items, err := result.Items()
	if err != nil {
		return nil, err
	}

	count, err := items.Size()
	if err != nil {
		return nil, err
	}

	history := make([]Item, 0, count)

	for i := uint32(0); i < count; i++ {
		item, err := items.GetAt(i)
		if err != nil {
			return nil, err
		}

		content, err := item.Content()
		if err != nil {
			return nil, err
		}

		formats, err := datatransfer.StandardDataFormatsStatics()
		if err != nil {
			return nil, err
		}

		textFormat, err := formats.Text()
		if err != nil {
			return nil, err
		}

		hasText, err := content.Contains(textFormat)
		if err != nil {
			return nil, err
		}

		if !hasText {
			continue
		}

		textOperation, err := content.GetTextAsync()
		if err != nil {
			return nil, err
		}

		id, err := item.Id()
		if err != nil {
			return nil, err
		}

		text, err := textOperation.Await()
		if err != nil {
			return nil, err
		}

		history = append(history, Item{
			ID:   id,
			Text: text,
		})
	}

	return history, nil
}

func Yank(id string) error {
	clipboard, err := datatransfer.ClipboardStatics2()
	if err != nil {
		return err
	}

	operation, err := clipboard.GetHistoryItemsAsync()
	if err != nil {
		return err
	}

	result, err := operation.Await()
	if err != nil {
		return err
	}

	items, err := result.Items()
	if err != nil {
		return err
	}

	count, err := items.Size()
	if err != nil {
		return err
	}

	for i := uint32(0); i < count; i++ {
		item, err := items.GetAt(i)
		if err != nil {
			return err
		}

		itemID, err := item.Id()
		if err != nil {
			return err
		}

		if itemID != id {
			continue
		}

		status, err := clipboard.SetHistoryItemAsContent(item)
		if err != nil {
			return err
		}

		if status != datatransfer.SetHistoryItemAsContentStatusSuccess {
			return fmt.Errorf("set history item failed: %s", status)
		}

		return nil
	}

	return errors.New("clipboard history item not found")
}
