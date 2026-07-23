import { render, screen, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect } from 'vitest';
import { ChatComposer } from './ChatComposer';

describe('ChatComposer', () => {
  it('reports typing and sends on click', () => {
    const onChange = vi.fn();
    const onSend = vi.fn();
    render(
      <ChatComposer
        value=""
        disabled={false}
        working={false}
        onChange={onChange}
        onSend={onSend}
        onStop={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByPlaceholderText('Ask Charli…'), { target: { value: 'x' } });
    expect(onChange).toHaveBeenCalledWith('x');

    fireEvent.click(screen.getByTitle('Send'));
    expect(onSend).toHaveBeenCalledTimes(1);
  });

  it('sends on Enter', () => {
    const onSend = vi.fn();
    render(
      <ChatComposer
        value="hi"
        disabled={false}
        working={false}
        onChange={() => {}}
        onSend={onSend}
        onStop={vi.fn()}
      />,
    );
    fireEvent.keyDown(screen.getByPlaceholderText('Ask Charli…'), { key: 'Enter' });
    expect(onSend).toHaveBeenCalled();
  });

  it('disables the send button while sending', () => {
    render(
      <ChatComposer
        value="hi"
        disabled
        working={false}
        onChange={() => {}}
        onSend={() => {}}
        onStop={vi.fn()}
      />,
    );
    expect(screen.getByTitle('Send')).toBeDisabled();
  });

  it('shows a stop button instead of send while working, and it calls onStop', () => {
    const onStop = vi.fn();
    render(
      <ChatComposer
        value="hi"
        disabled
        working
        onChange={() => {}}
        onSend={() => {}}
        onStop={onStop}
      />,
    );
    expect(screen.queryByTitle('Send')).not.toBeInTheDocument();
    fireEvent.click(screen.getByTitle('Stop (Esc)'));
    expect(onStop).toHaveBeenCalledTimes(1);
  });
});
