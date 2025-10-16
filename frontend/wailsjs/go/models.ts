export namespace model {
	
	export class Match {
	    match_id: number[];
	    round_number: number;
	    table_number: number;
	    player_a_id: string;
	    player_b_id: string;
	    white_id: string;
	    black_id: string;
	    result: string;
	    score_a: number;
	    score_b: number;
	
	    static createFrom(source: any = {}) {
	        return new Match(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.match_id = source["match_id"];
	        this.round_number = source["round_number"];
	        this.table_number = source["table_number"];
	        this.player_a_id = source["player_a_id"];
	        this.player_b_id = source["player_b_id"];
	        this.white_id = source["white_id"];
	        this.black_id = source["black_id"];
	        this.result = source["result"];
	        this.score_a = source["score_a"];
	        this.score_b = source["score_b"];
	    }
	}
	export class Player {
	    id: string;
	    name: string;
	    score: number;
	    opponent_ids: string[];
	    buchholz: number;
	    progressive_score: number;
	    head_to_head_results: Record<string, number>;
	    color_history: string;
	    has_bye: boolean;
	    club?: string;
	
	    static createFrom(source: any = {}) {
	        return new Player(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.score = source["score"];
	        this.opponent_ids = source["opponent_ids"];
	        this.buchholz = source["buchholz"];
	        this.progressive_score = source["progressive_score"];
	        this.head_to_head_results = source["head_to_head_results"];
	        this.color_history = source["color_history"];
	        this.has_bye = source["has_bye"];
	        this.club = source["club"];
	    }
	}
	export class Round {
	    round_number: number;
	    matches: Match[];
	    is_complete: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Round(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.round_number = source["round_number"];
	        this.matches = this.convertValues(source["matches"], Match);
	        this.is_complete = source["is_complete"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Tournament {
	    id: number[];
	    title: string;
	    description: string;
	    status: string;
	    players_data: number[];
	    rounds_data: number[];
	    events_data: number[];
	    current_round: number;
	    total_players: number;
	    // Go type: time
	    start_time: any;
	    // Go type: time
	    end_time?: any;
	    rounds_total?: number;
	    bye_score?: number;
	    pairing_system?: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Tournament(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.status = source["status"];
	        this.players_data = source["players_data"];
	        this.rounds_data = source["rounds_data"];
	        this.events_data = source["events_data"];
	        this.current_round = source["current_round"];
	        this.total_players = source["total_players"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.end_time = this.convertValues(source["end_time"], null);
	        this.rounds_total = source["rounds_total"];
	        this.bye_score = source["bye_score"];
	        this.pairing_system = source["pairing_system"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

