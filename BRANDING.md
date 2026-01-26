# Chimborazo: Narrative Branding Guide

> *"It is far more difficult to observe correctly than most men imagine; to behold is not necessarily to observe, and the power of comparing and combining is only to be obtained by education."*
>
> — Alexander von Humboldt

---

## The Narrative Voice

Chimborazo documentation is written from the perspective of **Alexander von Humboldt** in Paris, circa 1805-1834, as he processes and publishes the results of his five-year expedition to the Americas (1799-1804).

This is not Humboldt in the field—sweating in the Orinoco, gasping on Chimborazo's slopes. This is Humboldt *returned*: reflective, synthesizing, drawing connections between observations that seemed isolated in the moment. He is writing the 30 volumes that will occupy the rest of his life.

### Voice Characteristics

| Quality | Description |
|---------|-------------|
| **Retrospective** | Looking back on the journey, finding meaning |
| **Synthesizing** | Connecting disparate observations into patterns |
| **Systems thinking** | Everything is interconnected |
| **Scientific yet poetic** | Precise but not dry; wonder is permitted |
| **Generous** | Credits fellow travelers and predecessors freely |

### What the Voice Is NOT

- Not instructional or tutorial-like ("Step 1: Do this...")
- Not casual or modern ("Let's go ahead and...")
- Not self-aggrandizing (despite his achievements, he credits others)
- Not dry academic prose (he was a gifted writer)

---

## The Expedition Metaphor

The central metaphor is the **expedition**—specifically Humboldt's 1799-1804 journey through Venezuela, Cuba, Colombia, Ecuador, Peru, Mexico, and the United States, culminating in the 1802 ascent of Chimborazo.

### The Stages

| Stage | Expedition Phase | Code Phase | Location |
|-------|------------------|------------|----------|
| **Preparation** | Madrid: instruments, permissions, planning | Recipe parsing, validation | `internal/config/` |
| **Acquisition** | Cumaná to Quito: collecting specimens | Data fetching, caching | `internal/sources/` |
| **Transformation** | Comparing, combining, finding patterns | Geometry operations | `internal/geometry/` |
| **Revelation** | The Naturgemälde: visualization that reveals | SVG rendering | `internal/output/` |
| **Return** | Paris: 30 volumes synthesizing everything | Pipeline orchestration | `pkg/pipeline/` |

### The Summit

Chimborazo itself—19,286 feet, world altitude record 1802, stopped 1,000 feet from summit by an impassable crevasse—represents the asymptotic nature of cartographic work. We approach perfection; we never quite reach it. The view from 19,286 feet is still worth the climb.

---

## The Cartographic Ensemble

Humboldt does not work alone. He references five historical figures whose methods and principles inform the work. These are not package names—they are intellectual companions, predecessors whose wisdom echoes through the documentation.

### William Smith (1769-1839)
**"The Father of English Geology"**

| Detail | Value |
|--------|-------|
| **Era** | 1815 |
| **Achievement** | First geological map of a nation |
| **Principle** | Layer order matters |
| **Voice quality** | Practical, working-class, hard-won wisdom |

**Key Quote:**
> "The same strata were found always in the same order of superposition and contained the same peculiar fossils."

**In Chimborazo context:**
Humboldt references Smith when discussing layer ordering in SVG output—the importance of z-index, of placing backgrounds below foregrounds, of respecting the natural order of visual elements.

---

### Henry David Thoreau (1817-1862)
**Surveyor of Walden Pond**

| Detail | Value |
|--------|-------|
| **Era** | 1854 |
| **Achievement** | Precise surveys of Walden Pond; 200 professional surveys |
| **Principle** | Go to the source |
| **Voice quality** | Precise, observational, skeptical of hearsay |

**Key Quote:**
> "It is remarkable how long men will believe in the bottomlessness of a pond without taking the trouble to sound it."

**In Chimborazo context:**
Humboldt references Thoreau when discussing data acquisition—the importance of fetching authoritative sources rather than relying on speculation or cached assumptions. The `sources/` package embodies Thoreau's insistence on direct measurement.

---

### Alexander von Humboldt (1769-1859)
**"The Father of Ecology"**

| Detail | Value |
|--------|-------|
| **Era** | 1799-1859 |
| **Achievement** | Invented isotherms, biogeography, thematic mapping |
| **Principle** | Unity in diversity; everything is interconnected |
| **Voice quality** | Systems thinking, poetic science, generous to predecessors |

**Key Quotes:**
> "Nature considered rationally, that is to say, submitted to the process of thought, is a unity in diversity of phenomena; a harmony, blending together all created things."

> "If instead of geographic maps, we only possessed tables covering latitude, longitude, and altitude, a great number of curious connections that continents manifest in their forms would have stayed forever lost."

**In Chimborazo context:**
Humboldt IS the narrator. His voice frames all documentation. His transformation philosophy—comparing and combining to reveal hidden patterns—is the heart of the `geometry/` package.

---

### Matthew Fontaine Maury (1806-1873)
**"Pathfinder of the Seas"**

| Detail | Value |
|--------|-------|
| **Era** | 1855 |
| **Achievement** | Aggregated global ocean data from ship captains; first oceanography textbook |
| **Principle** | Coordinate many sources to reveal patterns invisible to individuals |
| **Voice quality** | Coordinating, synthesizing, pattern-recognizing |

**Key Quote:**
> "There is a river in the ocean. In the severest droughts it never fails, and in the mightiest floods it never overflows."

**In Chimborazo context:**
Humboldt references Maury when discussing pipeline orchestration—how coordinating data from Census, Natural Earth, OpenStreetMap, and other sources produces maps that no single source could yield. The `pipeline/` package embodies Maury's aggregation philosophy.

---

### Florence Kelley (1859-1932)
**Pioneer of Social Mapping**

| Detail | Value |
|--------|-------|
| **Era** | 1895 |
| **Achievement** | Hull House Maps and Papers; first systematic spatial investigation of an American neighborhood |
| **Principle** | Maps are evidence, not decoration |
| **Voice quality** | Direct, purposeful, accessible, evidence-based |

**Key Quotes:**
> "So that the plain man who pays for the census gets what he pays for, statistical data in a form in which they are ready for him to understand."

> "By looking at just eight pages, the reader could easily see the clustering of nationalities and the distribution of families' weekly wages."

**In Chimborazo context:**
Humboldt references Kelley when discussing output and visualization—SVG that serves comprehension, not mere aesthetics. Every style choice must make patterns visible that would otherwise stay hidden. The `output/` package embodies Kelley's evidence-based visualization philosophy.

---

## Writing Guidelines

### For Package Documentation (doc.go)

Open with a Humboldt-voice paragraph connecting the stage to the expedition, then transition to technical documentation. Example structure:

```go
/*
Package sources implements data acquisition from authoritative geospatial sources.

I learned in Cumaná what the surveyor Thoreau knew: one must go to the source.
It is remarkable how long cartographers will rely on cached assumptions without
taking the trouble to fetch the authoritative data. This package ensures that
every coordinate, every boundary, every water feature comes directly from the
agencies that surveyed them.

# Supported URI Schemes

The fetcher resolves URIs with the following schemes:
  - census://  - U.S. Census Bureau TIGER/Line files
  - nhn://     - Canadian National Hydro Network
  - file://    - Local files

# Caching

Downloaded files are cached in ~/.cache/chimborazo/...
*/
package sources
```

### For Error Messages

Keep them practical but not robotic. Humboldt would not write "ERROR: file not found." He might write:

```
failed to acquire source "census://water/VT": the Census Bureau returned no data for this region
```

### For README and User-Facing Docs

The expedition metaphor is primary. Users are fellow explorers following the same path.

### For Code Comments

These can be more technical—Humboldt did not narrate his code. Use the narrative voice for package-level and public API documentation; use standard Go idioms for implementation comments.

---

## Historical Assets

The following public domain materials can illustrate documentation:

| Image | Subject | Source |
|-------|---------|--------|
| Chimborazo cross-section | Humboldt's Naturgemälde | Internet Archive |
| Smith's 1815 geological map | England's strata | Natural History Museum |
| Hull House nationality map | Chicago's 19th Ward | Internet Archive |
| Thoreau's Walden survey | Pond with soundings | Walden Woods Project |
| Maury's Gulf Stream chart | Atlantic currents | LOC |

---

## Sample Prose

### Technical Introduction with Narrative

> The `geometry` package transforms raw coordinates into meaningful shapes. When we subtract water from land, we do not merely perform a Boolean operation on polygons—we reveal the true relationship between lake and shore, the way Champlain carves its space from Vermont and New York alike. When we merge the scattered census blocks of a county, we see the administrative boundary emerge from its fragments.
>
> I call this stage **Transformation** because it is here that unity emerges from diversity. Isolated observations become patterns. Coordinates become connections.

### Error Documentation with Narrative

> If the recipe references a source that cannot be found, the build will halt at the Acquisition stage. One cannot transform what one has not acquired. The error message will indicate which source failed and what the fetcher attempted—whether the Census Bureau was unreachable, or the URI malformed, or the cache corrupted.

---

## Design Principles (Derived from the Ensemble)

1. **Layer order matters** (Smith) — Visual stacking follows logical hierarchy
2. **Go to the source** (Thoreau) — Authoritative data, not speculation
3. **Unity in diversity** (Humboldt) — Transformation reveals connections
4. **Coordinate many sources** (Maury) — No single dataset tells the complete story
5. **Maps are evidence** (Kelley) — Output serves comprehension, not decoration

---

## The Name

**Chimborazo** is the Ecuadorian volcano whose peak is the farthest point from Earth's center (due to equatorial bulge). Humboldt's 1802 ascent reached 19,286 feet—a world altitude record that stood for 30 years.

The name was chosen because:
1. It directly references Humboldt's iconic achievement
2. It avoids naming conflicts with other GIS tools
3. It embodies the project's goals: reaching heights that reveal new perspectives
4. The Naturgemälde diagram of Chimborazo's vegetation zones is the original "layered data visualization"

---

> *"I have ever desired to discern spatial phenomena in their widest mutual connection, and to comprehend Geography as a whole, animated and moved by inward forces."*
>
> — Humboldt
