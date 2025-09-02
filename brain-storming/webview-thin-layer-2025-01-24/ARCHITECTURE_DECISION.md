# Architektur-Entscheidungen

## Finale Architektur

### Stack-Entscheidungen

| Komponente | Gewählt | Alternative | Begründung |
|------------|---------|-------------|------------|
| **GUI Framework** | webview/webview | Wails v2 | 60% kleiner, minimaler |
| **DOM Updates** | idiomorph | Full Replace | Smooth In-State Updates |
| **Template Engine** | Scriggo | html/template | Bereits im Einsatz |
| **State Management** | DFA | Switch/Case | Testbar, erweiterbar |
| **Data Flow** | Real-Time Sync | Batch Collection | Einfacher, sauberer |

## Architektur-Prinzipien

### 1. Thin-Layer WebView
- WebView ist nur ein "HTML Viewer"
- Keine Business Logic im Frontend
- Minimales JavaScript (< 50 LOC)

### 2. DFA-First
- State Machine als Fundament
- Vollständig ohne UI testbar
- Business Logic in Go

### 3. Server-Side Rendering
- Scriggo generiert HTML
- Templates pro State
- Keine Client-Side Frameworks

### 4. Two-Level Updates
- **Level 1**: State Transitions → Full Replace
- **Level 2**: In-State Changes → idiomorph

### 5. Real-Time Data Sync
- Jede Änderung sofort an Go
- Kein "Submit" am Ende nötig
- DFA hat immer aktuellen State

## Datenfluss

```
1. User ändert Checkbox
   ↓
2. onChange Event
   ↓
3. update(field, value) → Go
   ↓
4. DFA.SetData(field, value)
   ↓
5. Scriggo rendert betroffene Bereiche
   ↓
6. idiomorph.morph(oldDOM, newHTML)
   ↓
7. UI updated smooth
```

## Code-Struktur

```
setupkit/
├── cmd/
│   └── installer/
│       └── main.go           # webview initialization
├── pkg/
│   ├── wizard/
│   │   ├── dfa.go           # State Machine
│   │   └── providers/        # DFA Providers
│   └── ui/
│       ├── bridge.go        # webview Bridge
│       └── renderer.go      # Scriggo SSR
└── templates/
    ├── layouts/
    │   └── base.html        # Mit idiomorph
    ├── states/
    │   ├── welcome.html
    │   ├── license.html
    │   └── components.html
    └── partials/
        └── component-list.html
```

## JavaScript Minimierung

```javascript
// DAS ist das GESAMTE Frontend JavaScript:
window.setupkit = {
    // Real-time update
    update: (field, value) => {
        window.update(field, value, {morph: true});
    },
    
    // State navigation  
    navigate: (action) => {
        window.navigate(action);
    }
};

// Event handlers direkt in HTML:
<input onchange="setupkit.update('license_accepted', this.checked)">
<button onclick="setupkit.navigate('next')">Next</button>
```

## Binary-Zusammensetzung

```
webview base:          4-5 MB
WebView runtime:       3-4 MB
Go Code:              2-3 MB
Embedded Templates:    0.5 MB
idiomorph (embedded):  0.006 MB
───────────────────────────────
TOTAL:                ~10-12 MB
```

## Entscheidungsmatrix

| Kriterium | Wails | webview | Gewinner |
|-----------|-------|---------|----------|
| Binary Size | 25-30 MB | 10-12 MB | ✅ webview |
| Complexity | Hoch | Niedrig | ✅ webview |
| Features für Thin-Layer | 10% genutzt | 100% genutzt | ✅ webview |
| Maintenance | Framework-Updates | Minimal | ✅ webview |
| Dev Experience | Gut | Ausreichend | ⚠️ Wails |
| Control | Black Box | Transparent | ✅ webview |

## Migration Plan

### Phase 1: Proof of Concept (1 Tag)
- webview Hello World
- DFA Integration Test
- idiomorph Test

### Phase 2: Implementation (2 Tage)
- Bridge Functions
- SSR Integration
- State Templates

### Phase 3: Testing (1 Tag)
- Cross-Platform Tests
- Performance Messung
- Binary Size Optimierung

## Risiko-Bewertung

| Risiko | Wahrscheinlichkeit | Impact | Mitigation |
|--------|-------------------|--------|------------|
| webview Bugs | Niedrig | Mittel | Aktives Projekt, Community |
| Keine Hot-Reload | Sicher | Niedrig | SSR ist schnell genug |
| Platform-Unterschiede | Mittel | Niedrig | Testing auf allen OS |

## Finale Empfehlung

✅ **webview/webview implementieren**

Die Vorteile überwiegen deutlich:
- 60% kleinere Binary
- Drastisch reduzierte Komplexität
- Perfekt für Thin-Layer Ansatz
- Volle Kontrolle über jeden Aspekt

Wails war die richtige Wahl für den Start, aber mit dem klaren Thin-Layer Konzept ist webview die logische Evolution.