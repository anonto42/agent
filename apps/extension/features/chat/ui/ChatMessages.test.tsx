import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { ChatMessages } from './ChatMessages';

describe('ChatMessages', () => {
  it('shows the empty state when there are no messages', () => {
    render(<ChatMessages messages={[]} />);
    expect(screen.getByText(/ask charli anything/i)).toBeInTheDocument();
  });

  it('renders user and assistant messages', () => {
    render(
      <ChatMessages
        messages={[
          { role: 'user', content: 'hello' },
          { role: 'assistant', content: 'hi there' },
        ]}
      />,
    );
    expect(screen.getByText('hello')).toBeInTheDocument();
    expect(screen.getByText('hi there')).toBeInTheDocument();
  });
});
