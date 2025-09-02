# Go GUI Frameworks - Vergleichsreport 2025

## Executive Summary

Wails ist aktuell die modernste Lösung für Desktop-Apps mit Web-Technologien. Alternativen wie Fyne oder Gio bieten native Performance, aber weniger Flexibilität im UI-Design. Die Binary-Größen variieren stark von 8 MB (Gio) bis 150 MB (Electron-basierte Lösungen).

## Größenvergleich der Frameworks

| Framework | Minimale App | Typische App | Große App | Runtime-Deps |
|-----------|--------------|--------------|-----------|--------------|
| **Wails v2** | 10-15 MB | 20-30 MB | 40-60 MB | WebView2/WebKit |
| **Fyne** | 12-18 MB | 25-35 MB | 45-70 MB | OpenGL |
| **Gio** | 8-12 MB | 15-25 MB | 30-50 MB | Keine |
| **Walk (Windows)** | 5-8 MB | 10-15 MB | 20-30 MB | Win32 API |
| **Electron + Go** | 80-100 MB | 120-150 MB | 200+ MB | Chromium |
| **Tauri + Go** | 8-12 MB | 15-25 MB | 30-50 MB | WebView2/WebKit |
| **Qt (therecipe/qt)** | 50-70 MB | 80-100 MB | 150+ MB | Qt Libraries |
| **GTK (gotk3)** | 20-30 MB | 35-50 MB | 60-80 MB | GTK3 |
| **Webview/webview** | 8-10 MB | 12-18 MB | 25-35 MB | System WebView |
| **Lorca** | 10-12 MB | 15-20 MB | 30-40 MB | Chrome/Chromium |

## Performance-Vergleich

| Framework | Startup Zeit | RAM (Idle) | RAM (Active) | CPU Usage | FPS (Animations) |
|-----------|--------------|------------|--------------|-----------|------------------|
| **Wails** | 200-400ms | 30-50 MB | 60-100 MB | Low | 60 FPS |
| **Fyne** | 150-300ms | 40-60 MB | 80-120 MB | Medium | 60 FPS |
| **Gio** | 100-200ms | 20-30 MB | 40-60 MB | Low | 60-120 FPS |
| **Walk** | 50-150ms | 15-25 MB | 30-50 MB | Very Low | N/A |
| **Electron** | 1-3s | 150-200 MB | 300-500 MB | High | 60 FPS |
| **Qt** | 300-500ms | 60-80 MB | 100-150 MB | Medium | 60 FPS |

## Feature-Matrix

| Feature | Wails | Fyne | Gio | Walk | Electron | Tauri | Qt | GTK |
|---------|-------|------|-----|------|----------|-------|----|----|
| Cross-Platform | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ⚠️ |
| Mobile Support | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ |
| Web Technologies | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ⚠️ | ❌ |
| Native Widgets | ⚠️ | ❌ | ❌ | ✅ | ❌ | ⚠️ | ✅ | ✅ |
| Custom Rendering | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ |
| Hot Reload | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ |
| Small Binary | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ⚠️ |
| No Dependencies | ❌ | ⚠️ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Theming | ✅ | ✅ | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ |
| Accessibility | ✅ | ⚠️ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |

## Empfehlung für SetupKit

### Warum webview/webview die bessere Wahl ist:

1. **Optimale Binary-Größe** (10-12 MB)
   - 60% kleiner als Wails
   - Perfekt für Installer

2. **Minimal API**
   - Genau was für Thin-Layer benötigt wird
   - Keine unnötigen Features

3. **Volle Kontrolle**
   - Transparente Implementation
   - Kein Framework-Overhead

4. **SetupKit-spezifische Vorteile**
   - SSR mit Scriggo passt perfekt
   - DFA-Integration nahtlos
   - Real-Time Data Sync möglich

## Migrations-Aufwand von Wails

| Ziel-Framework | Aufwand | Risiko | Lohnt sich? |
|----------------|---------|--------|-------------|
| **webview** | 2-3 Tage | Niedrig | ✅ Definitiv |
| **Fyne** | 3-4 Wochen | Mittel | ❌ Nur wenn Mobile wichtig |
| **Gio** | 6-8 Wochen | Hoch | ❌ Zu viel Aufwand |
| **Tauri** | 2-3 Wochen | Niedrig | ⚠️ Marginal besser |
| **Electron** | 1-2 Wochen | Niedrig | ❌ Größe inakzeptabel |

## Fazit

**webview/webview ist die optimale Wahl für SetupKit** weil:

1. **Balance**: Beste Balance zwischen Einfachheit und Features
2. **Größe**: Nur 10-12 MB Binary
3. **Wartbarkeit**: Minimaler Code, maximale Kontrolle
4. **Integration**: Perfekt für DFA + SSR Ansatz

Für einen **Installer** ist webview mit 10-12 MB die beste Wahl - klein genug zum Download, einfach genug zur Wartung.