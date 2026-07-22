import { render, screen, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect } from 'vitest';
import { ChatComposer } from './ChatComposer';

describe('ChatComposer', () => {
  it('reports typing and sends on click', () => {
    const onChange = vi.fn();
    const onSend = vi.fn();
    render(<ChatComposer value="" disabled={false} onChange={onChange} onSend={onSend} />);

    fireEvent.change(screen.getByPlaceholderText('Ask Charli…'), { target: { value: 'x' } });
    expect(onChange).toHaveBeenCalledWith('x');

    fireEvent.click(screen.getByTitle('Send'));
    expect(onSend).toHaveBeenCalledTimes(1);
  });

  it('sends on Enter', () => {
    const onSend = vi.fn();
    render(<ChatComposer value="hi" disabled={false} onChange={() => {}} onSend={onSend} />);
    fireEvent.keyDown(screen.getByPlaceholderText('Ask Charli…'), { key: 'Enter' });
    expect(onSend).toHaveBeenCalled();
  });

  it('disables the send button while sending', () => {
    render(<ChatComposer value="hi" disabled onChange={() => {}} onSend={() => {}} />);
    expect(screen.getByTitle('Send')).toBeDisabled();
  });
});
