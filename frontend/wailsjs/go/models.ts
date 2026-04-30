export namespace model {
	
	export class ApplyResult {
	    NewPath: string;
	    Moved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApplyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.NewPath = source["NewPath"];
	        this.Moved = source["Moved"];
	    }
	}
	export class FileInfo {
	    Path: string;
	    Name: string;
	    Type: string;
	    // Go type: time
	    ModTime: any;
	    // Go type: time
	    BirthTime: any;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.ModTime = this.convertValues(source["ModTime"], null);
	        this.BirthTime = this.convertValues(source["BirthTime"], null);
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
	export class Tag {
	    ID: number;
	    Name: string;
	    Type: string;
	    Folder: string;
	    Color: string;
	    Hotkey: string;
	
	    static createFrom(source: any = {}) {
	        return new Tag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Folder = source["Folder"];
	        this.Color = source["Color"];
	        this.Hotkey = source["Hotkey"];
	    }
	}

}

