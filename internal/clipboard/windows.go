package clipboard

import (
	"errors"
	"runtime"

	"github.com/deploymenttheory/go-bindings-winrt/bindings/winrt/applicationmodel/datatransfer"
	"golang.org/x/sys/windows"
)

var ErrHistoryDisabled = errors.New("Windows Clipboard History is disabled")

type Item struct {
	Text string
}

func History() ([]Item, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err != nil {
		return nil, err
	}
	defer windows.CoUninitialize()

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

		textOperation, err := content.GetTextAsync()
		if err != nil {
			return nil, err
		}

		text, err := textOperation.Await()
		if err != nil {
			return nil, err
		}

		history = append(history, Item{
			Text: text,
		})
	}

	return history, nil
}
