package arxiv

// Category represents arXiv categories
type Category string

const (
	// Computer Science
	CategoryCSAI Category = "cs.AI" // Artificial Intelligence
	CategoryCSAR Category = "cs.AR" // Hardware Architecture
	CategoryCSCC Category = "cs.CC" // Computational Complexity
	CategoryCSCE Category = "cs.CE" // Computational Engineering, Finance, and Science
	CategoryCSCG Category = "cs.CG" // Computational Geometry
	CategoryCSCL Category = "cs.CL" // Computation and Language
	CategoryCSCR Category = "cs.CR" // Cryptography and Security
	CategoryCSCV Category = "cs.CV" // Computer Vision and Pattern Recognition
	CategoryCSCY Category = "cs.CY" // Computers and Society
	CategoryCSDB Category = "cs.DB" // Databases
	CategoryCSDC Category = "cs.DC" // Distributed, Parallel, and Cluster Computing
	CategoryCSDL Category = "cs.DL" // Digital Libraries
	CategoryCSDM Category = "cs.DM" // Discrete Mathematics
	CategoryCSDS Category = "cs.DS" // Data Structures and Algorithms
	CategoryCSET Category = "cs.ET" // Emerging Technologies
	CategoryCSFL Category = "cs.FL" // Formal Languages and Automata Theory
	CategoryCSGL Category = "cs.GL" // General Literature
	CategoryCSGR Category = "cs.GR" // Graphics
	CategoryCSGT Category = "cs.GT" // Computer Science and Game Theory
	CategoryCSHC Category = "cs.HC" // Human-Computer Interaction
	CategoryCSIR Category = "cs.IR" // Information Retrieval
	CategoryCSIT Category = "cs.IT" // Information Theory
	CategoryCSLG Category = "cs.LG" // Machine Learning
	CategoryCSLO Category = "cs.LO" // Logic in Computer Science
	CategoryCSMA Category = "cs.MA" // Multiagent Systems
	CategoryCSMM Category = "cs.MM" // Multimedia
	CategoryCSMS Category = "cs.MS" // Mathematical Software
	CategoryCSNA Category = "cs.NA" // Numerical Analysis
	CategoryCSNE Category = "cs.NE" // Neural and Evolutionary Computing
	CategoryCSNI Category = "cs.NI" // Networking and Internet Architecture
	CategoryCSOH Category = "cs.OH" // Other Computer Science
	CategoryCSOS Category = "cs.OS" // Operating Systems
	CategoryCSPF Category = "cs.PF" // Performance
	CategoryCSPL Category = "cs.PL" // Programming Languages
	CategoryCSRO Category = "cs.RO" // Robotics
	CategoryCSSC Category = "cs.SC" // Symbolic Computation
	CategoryCSSD Category = "cs.SD" // Sound
	CategoryCSSE Category = "cs.SE" // Software Engineering
	CategoryCSSI Category = "cs.SI" // Social and Information Networks
	CategoryCSSY Category = "cs.SY" // Systems and Control

	// Economics
	CategoryEconEM Category = "econ.EM" // Econometrics
	CategoryEconGN Category = "econ.GN" // General Economics
	CategoryEconTH Category = "econ.TH" // Theoretical Economics

	// Electrical Engineering and Systems Science
	CategoryEESSAS Category = "eess.AS" // Audio and Speech Processing
	CategoryEESSIV Category = "eess.IV" // Image and Video Processing
	CategoryEESSSP Category = "eess.SP" // Signal Processing
	CategoryEESSSY Category = "eess.SY" // Systems and Control

	// Mathematics
	CategoryMathAC Category = "math.AC" // Commutative Algebra
	CategoryMathAG Category = "math.AG" // Algebraic Geometry
	CategoryMathAP Category = "math.AP" // Analysis of PDEs
	CategoryMathAT Category = "math.AT" // Algebraic Topology
	CategoryMathCA Category = "math.CA" // Classical Analysis and ODEs
	CategoryMathCO Category = "math.CO" // Combinatorics
	CategoryMathCT Category = "math.CT" // Category Theory
	CategoryMathCV Category = "math.CV" // Complex Variables
	CategoryMathDG Category = "math.DG" // Differential Geometry
	CategoryMathDS Category = "math.DS" // Dynamical Systems
	CategoryMathFA Category = "math.FA" // Functional Analysis
	CategoryMathGM Category = "math.GM" // General Mathematics
	CategoryMathGN Category = "math.GN" // General Topology
	CategoryMathGR Category = "math.GR" // Group Theory
	CategoryMathGT Category = "math.GT" // Geometric Topology
	CategoryMathHO Category = "math.HO" // History and Overview
	CategoryMathIT Category = "math.IT" // Information Theory
	CategoryMathKT Category = "math.KT" // K-Theory and Homology
	CategoryMathLO Category = "math.LO" // Logic
	CategoryMathMG Category = "math.MG" // Metric Geometry
	CategoryMathMP Category = "math.MP" // Mathematical Physics
	CategoryMathNA Category = "math.NA" // Numerical Analysis
	CategoryMathNT Category = "math.NT" // Number Theory
	CategoryMathOA Category = "math.OA" // Operator Algebras
	CategoryMathOC Category = "math.OC" // Optimization and Control
	CategoryMathPR Category = "math.PR" // Probability
	CategoryMathQA Category = "math.QA" // Quantum Algebra
	CategoryMathRA Category = "math.RA" // Rings and Algebras
	CategoryMathRT Category = "math.RT" // Representation Theory
	CategoryMathSG Category = "math.SG" // Symplectic Geometry
	CategoryMathSP Category = "math.SP" // Spectral Theory
	CategoryMathST Category = "math.ST" // Statistics Theory

	// Physics - Astrophysics
	CategoryAstroPh   Category = "astro-ph"    // Astrophysics (general)
	CategoryAstroPhCO Category = "astro-ph.CO" // Cosmology and Nongalactic Astrophysics
	CategoryAstroPhEP Category = "astro-ph.EP" // Earth and Planetary Astrophysics
	CategoryAstroPhGA Category = "astro-ph.GA" // Astrophysics of Galaxies
	CategoryAstroPhHE Category = "astro-ph.HE" // High Energy Astrophysical Phenomena
	CategoryAstroPhIM Category = "astro-ph.IM" // Instrumentation and Methods for Astrophysics
	CategoryAstroPhSR Category = "astro-ph.SR" // Solar and Stellar Astrophysics

	// Physics - Condensed Matter
	CategoryCondMat         Category = "cond-mat"           // Condensed Matter (general)
	CategoryCondMatDisNn    Category = "cond-mat.dis-nn"    // Disordered Systems and Neural Networks
	CategoryCondMatMesHall  Category = "cond-mat.mes-hall"  // Mesoscale and Nanoscale Physics
	CategoryCondMatMtrlSci  Category = "cond-mat.mtrl-sci"  // Materials Science
	CategoryCondMatOther    Category = "cond-mat.other"     // Other Condensed Matter
	CategoryCondMatQuantGas Category = "cond-mat.quant-gas" // Quantum Gases
	CategoryCondMatSoft     Category = "cond-mat.soft"      // Soft Condensed Matter
	CategoryCondMatStatMech Category = "cond-mat.stat-mech" // Statistical Mechanics
	CategoryCondMatStrEl    Category = "cond-mat.str-el"    // Strongly Correlated Electrons
	CategoryCondMatSuprCon  Category = "cond-mat.supr-con"  // Superconductivity

	// Physics - General Relativity and Quantum Cosmology
	CategoryGrQc Category = "gr-qc" // General Relativity and Quantum Cosmology

	// Physics - High Energy Physics
	CategoryHepEx  Category = "hep-ex"  // High Energy Physics - Experiment
	CategoryHepLat Category = "hep-lat" // High Energy Physics - Lattice
	CategoryHepPh  Category = "hep-ph"  // High Energy Physics - Phenomenology
	CategoryHepTh  Category = "hep-th"  // High Energy Physics - Theory

	// Physics - Mathematical Physics
	CategoryMathPh Category = "math-ph" // Mathematical Physics

	// Physics - Nonlinear Sciences
	CategoryNlinAO Category = "nlin.AO" // Adaptation and Self-Organizing Systems
	CategoryNlinCD Category = "nlin.CD" // Chaotic Dynamics
	CategoryNlinCG Category = "nlin.CG" // Cellular Automata and Lattice Gases
	CategoryNlinPS Category = "nlin.PS" // Pattern Formation and Solitons
	CategoryNlinSI Category = "nlin.SI" // Exactly Solvable and Integrable Systems

	// Physics - Nuclear Physics
	CategoryNuclEx Category = "nucl-ex" // Nuclear Experiment
	CategoryNuclTh Category = "nucl-th" // Nuclear Theory

	// Physics - General Physics
	CategoryPhysicsAccPh   Category = "physics.acc-ph"   // Accelerator Physics
	CategoryPhysicsAoPh    Category = "physics.ao-ph"    // Atmospheric and Oceanic Physics
	CategoryPhysicsAppPh   Category = "physics.app-ph"   // Applied Physics
	CategoryPhysicsAtmClus Category = "physics.atm-clus" // Atomic and Molecular Clusters
	CategoryPhysicsAtomPh  Category = "physics.atom-ph"  // Atomic Physics
	CategoryPhysicsBioPh   Category = "physics.bio-ph"   // Biological Physics
	CategoryPhysicsChemPh  Category = "physics.chem-ph"  // Chemical Physics
	CategoryPhysicsClassPh Category = "physics.class-ph" // Classical Physics
	CategoryPhysicsCompPh  Category = "physics.comp-ph"  // Computational Physics
	CategoryPhysicsDataAn  Category = "physics.data-an"  // Data Analysis, Statistics and Probability
	CategoryPhysicsEdPh    Category = "physics.ed-ph"    // Physics Education
	CategoryPhysicsFluDyn  Category = "physics.flu-dyn"  // Fluid Dynamics
	CategoryPhysicsGenPh   Category = "physics.gen-ph"   // General Physics
	CategoryPhysicsGeoPh   Category = "physics.geo-ph"   // Geophysics
	CategoryPhysicsHistPh  Category = "physics.hist-ph"  // History and Philosophy of Physics
	CategoryPhysicsInsDet  Category = "physics.ins-det"  // Instrumentation and Detectors
	CategoryPhysicsMedPh   Category = "physics.med-ph"   // Medical Physics
	CategoryPhysicsOptics  Category = "physics.optics"   // Optics
	CategoryPhysicsPlasmPh Category = "physics.plasm-ph" // Plasma Physics
	CategoryPhysicsPopPh   Category = "physics.pop-ph"   // Popular Physics
	CategoryPhysicsSocPh   Category = "physics.soc-ph"   // Physics and Society
	CategoryPhysicsSpacePh Category = "physics.space-ph" // Space Physics

	// Physics - Quantum Physics
	CategoryQuantPh Category = "quant-ph" // Quantum Physics

	// Quantitative Biology
	CategoryQBioBM Category = "q-bio.BM" // Biomolecules
	CategoryQBioCB Category = "q-bio.CB" // Cell Behavior
	CategoryQBioGN Category = "q-bio.GN" // Genomics
	CategoryQBioMN Category = "q-bio.MN" // Molecular Networks
	CategoryQBioNC Category = "q-bio.NC" // Neurons and Cognition
	CategoryQBioOT Category = "q-bio.OT" // Other Quantitative Biology
	CategoryQBioPE Category = "q-bio.PE" // Populations and Evolution
	CategoryQBioQM Category = "q-bio.QM" // Quantitative Methods
	CategoryQBioSC Category = "q-bio.SC" // Subcellular Processes
	CategoryQBioTO Category = "q-bio.TO" // Tissues and Organs

	// Quantitative Finance
	CategoryQFinCP Category = "q-fin.CP" // Computational Finance
	CategoryQFinEC Category = "q-fin.EC" // Economics
	CategoryQFinGN Category = "q-fin.GN" // General Finance
	CategoryQFinMF Category = "q-fin.MF" // Mathematical Finance
	CategoryQFinPM Category = "q-fin.PM" // Portfolio Management
	CategoryQFinPR Category = "q-fin.PR" // Pricing of Securities
	CategoryQFinRM Category = "q-fin.RM" // Risk Management
	CategoryQFinST Category = "q-fin.ST" // Statistical Finance
	CategoryQFinTR Category = "q-fin.TR" // Trading and Market Microstructure

	// Statistics
	CategoryStatAP Category = "stat.AP" // Applications
	CategoryStatCO Category = "stat.CO" // Computation
	CategoryStatME Category = "stat.ME" // Methodology
	CategoryStatML Category = "stat.ML" // Machine Learning
	CategoryStatOT Category = "stat.OT" // Other Statistics
	CategoryStatTH Category = "stat.TH" // Statistics Theory
)

// SortCriterion represents sort criteria for search results
type SortCriterion string

const (
	SortByRelevance       SortCriterion = "relevance"
	SortByLastUpdatedDate SortCriterion = "lastUpdatedDate"
	SortBySubmittedDate   SortCriterion = "submittedDate"
)

// SortOrder represents sort order for search results
type SortOrder string

const (
	SortOrderAscending  SortOrder = "ascending"
	SortOrderDescending SortOrder = "descending"
)
