<p align="center">
  <img src="public/pmtop.png" width="400" alt="pmtop logo">
  <br>
  <b>pmtop</b>
</p>

`pmtop` is a tool designed to help you monitor and manage network ports on your system. Unlike traditional tools, `pmtop` shows you exactly which apps and project directories are using your ports, making it easy to identify and close unwanted processes.

## Installation

To install or update `pmtop` instantly on Windows, run the following command in **PowerShell**:

```powershell
irm https://raw.githubusercontent.com/dennismutuku2005/pmtop-lite/main/install.ps1 | iex
```

> [!IMPORTANT]
> **Run as Administrator**: To scan all system ports and safely close processes, please run your terminal or the `pmtop-gui` app as an Administrator.

## Features

- **Service & Framework Detection**: Automatically identifies **Node.js**, **Next.js**, **MySQL**, **Oracle**, **Redis**, and more based on port and process behavior.
- **Modern Dashboard GUI**: A light-themed interface with sidebar filters and overview cards.
- **Port Activity Logger**: Track every open, close, and process change on any port in real-time.

## Using the Terminal

Once installed, you can use `pmtop` in your terminal:

- **`pmtop`**: Launches the interactive dashboard.
- **`pmtop close <port>`**: Safely closes the process using a specific port (e.g., `pmtop close 3000`).
- **`pmtop log <port>`**: Starts a real-time activity logger for a specific port.
- **`pmtop --all`**: Shows all active connections, including established ones.

## Using the Desktop App

Search for **"pmtop"** in your Windows Start Menu to launch the dashboard. Use the search bar to find ports, and click on any row to see details or close the process.

## Bugs & Support

If you find a bug or have a feature request, please [create an issue](https://github.com/dennismutuku2005/pmtop-lite/issues) on GitHub.

---
© 2026 pmtop Team.