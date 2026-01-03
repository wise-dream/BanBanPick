export type MapName = 
  | "Ascent" 
  | "Bind" 
  | "Haven" 
  | "Split" 
  | "Lotus" 
  | "Icebox"
  | "Sunset" 
  | "Pearl" 
  | "Fracture" 
  | "Breeze" 
  | "Corrode" 
  | "Abyss";

export type Team = "A" | "B";

export interface VetoState {
  currentTeam: Team;
  bans: MapName[];
  pickedMap: MapName | null;
  finished: boolean;
  started: boolean;
}

export interface LogEntry {
  message: string;
  timestamp: number;
}
