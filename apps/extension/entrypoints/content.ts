// Content script: Charli's eyes and hands on the page.
//   PERCEPTION -> read the accessibility tree / selection / visible elements
//   ACTION     -> click, type, fill (gated by the backend safety engine)
export default defineContentScript({
  matches: ['<all_urls>'],
  main() {
    console.log('[charli] content script injected');
  },
});
