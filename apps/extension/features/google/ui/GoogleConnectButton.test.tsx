import { render, screen, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect } from 'vitest';

const mockUseGoogleConnection = vi.fn();
vi.mock('../hooks/useGoogleConnection', () => ({
  useGoogleConnection: () => mockUseGoogleConnection(),
}));

import { GoogleConnectButton } from './GoogleConnectButton';

describe('GoogleConnectButton', () => {
  it('shows a connected badge when connected', () => {
    mockUseGoogleConnection.mockReturnValue({ connected: true, connecting: false, connect: vi.fn() });
    render(<GoogleConnectButton />);
    expect(screen.getByText('Sheets ✓')).toBeInTheDocument();
  });

  it('shows a connect button and calls connect() on click when not connected', () => {
    const connect = vi.fn();
    mockUseGoogleConnection.mockReturnValue({ connected: false, connecting: false, connect });
    render(<GoogleConnectButton />);

    const button = screen.getByTitle('Connect Google Sheets');
    expect(button).toHaveTextContent('Connect Sheets');
    fireEvent.click(button);
    expect(connect).toHaveBeenCalledTimes(1);
  });

  it('disables the button and shows a connecting label while connecting', () => {
    mockUseGoogleConnection.mockReturnValue({ connected: false, connecting: true, connect: vi.fn() });
    render(<GoogleConnectButton />);

    const button = screen.getByTitle('Connect Google Sheets');
    expect(button).toBeDisabled();
    expect(button).toHaveTextContent('Connecting…');
  });
});
