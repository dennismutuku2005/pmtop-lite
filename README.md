<p align="center">
  <img src="public/pmtop.png" width="400" alt="pmtop logo">
  <br>
  <b>pmtop</b>
</p>

## Installation

To install or update the full `pmtop` suite (CLI + GUI), run this command in **PowerShell**:

```powershell
irm https://raw.githubusercontent.com/dennismutuku2005/pmtop-lite/main/install.ps1 | iex
```

> [!TIP]
> The installer now automatically detects running versions of pmtop and safely closes them to perform a clean update. No manual cleanup needed!

## The Developer Intelligence Engine

`pmtop` is built for developers. It doesn't just show ports; it understands your stack:

- **Auto-Detection**: Instantly recognizes **Oracle**, **Java/Tomcat**, **MySQL**, **Node.js**, **Python**, **PHP**, and **Go**.
- **Noise Suppression**: By default, it silences "Non-Dev" traffic. No more clutter from **Chrome**, **Slack**, **Discord**, or **OneDrive**.
- **Dev Mode Toggle**: Switch between a laser-focused "Clean View" and a full "System View" with one click (GUI) or by pressing `a` (CLI).

## Modern Interfaces

### 🎨 Material Desktop GUI
A premium desktop experience built for high productivity:
- **Tabbed Layout**: Dedicated **Dashboard** and **Settings** pages.
- **Status Indicators**: Live **● Green/Orange** dots show service health at a glance.
- **Performance tracking**: Real-time memory usage and uptime for every service.

### ⌨️ Redesigned CLI Dashboard
A high-performance terminal UI for the command-line power user:
- **System Header**: Real-time CPU and Memory bars (HTOP-style) right in your terminal.
- **Enhanced Visibility**: Clean, responsive table layout with colored status indicators.

## Usage

- **`pmtop`**: Launch the terminal dashboard.
- **`pmtop-gui`**: Search "pmtop" in your Start Menu for the desktop experience.
- **Admin Rights**: For full detection (including MySQL/Apache on Windows), please run as **Administrator**.

## Bugs & Support

If you find a bug or have a feature request, please [create an issue](https://github.com/dennismutuku2005/pmtop-lite/issues) on GitHub.

---
© 2026 pmtop Team.