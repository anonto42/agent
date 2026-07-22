// Background service worker: message router + backend connection.
// This is where the websocket client to the Go backend will live.
export default defineBackground(() => {
  // Open the side panel when Charli's toolbar icon is clicked.
  chrome.action?.onClicked.addListener((tab) => {
    if (tab.windowId != null) {
      chrome.sidePanel?.open({ windowId: tab.windowId });
    }
  });

  console.log('[charli] background ready');
});
