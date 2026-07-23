import { describe, it, expect } from 'vitest';
import { domFillFirstTextInput, domClickByText } from './domActions';

describe('domFillFirstTextInput', () => {
  it('fills the first text input and fires input/change events', () => {
    document.body.innerHTML = '<input type="text" id="a" />';
    const input = document.getElementById('a') as HTMLInputElement;
    let changed = false;
    input.addEventListener('change', () => (changed = true));

    const ok = domFillFirstTextInput('hello');

    expect(ok).toBe(true);
    expect(input.value).toBe('hello');
    expect(changed).toBe(true);
  });

  it('prefers a textarea when no text input exists', () => {
    document.body.innerHTML = '<textarea id="t"></textarea>';
    expect(domFillFirstTextInput('hi')).toBe(true);
    expect((document.getElementById('t') as HTMLTextAreaElement).value).toBe('hi');
  });

  it('returns false when there is nothing to fill', () => {
    document.body.innerHTML = '<div>no inputs here</div>';
    expect(domFillFirstTextInput('hi')).toBe(false);
  });
});

describe('domClickByText', () => {
  it('clicks the button whose text matches (case-insensitive, partial)', () => {
    document.body.innerHTML = '<button id="b">Submit Form</button>';
    let clicked = false;
    document.getElementById('b')!.addEventListener('click', () => (clicked = true));

    expect(domClickByText('submit')).toBe(true);
    expect(clicked).toBe(true);
  });

  it('matches links and role=button elements too', () => {
    document.body.innerHTML = '<a href="#">Next page</a>';
    expect(domClickByText('next')).toBe(true);
  });

  it('returns false when nothing matches', () => {
    document.body.innerHTML = '<button>Cancel</button>';
    expect(domClickByText('submit')).toBe(false);
  });
});
