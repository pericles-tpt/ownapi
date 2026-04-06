// NOTE: The structure below is translated from go structure on the backend
export type WhereClause = WhereGroup

export type WhereGroup = {
	groupOp: GroupOp
	groups:     WhereGroup[] | null
	conditions: WhereCondition[] | null 
}

export enum GroupOp {
	Not = "NOT",
	Or  = "OR",
	And = "AND",
}

/*
Represents a "simple" condition in a where clause
*/
export type WhereCondition = {
	fieldName: string
	op: ConditionOp
	fieldValue: string
	isStringLiteral: boolean
}

export enum ConditionOp {
	Equal            = "=",
	GreaterThan      = ">",
	LessThan         = "<",
	GreaterThanEqual = ">=",
	LessThanEqual    = "<=",
	NotEqual         = "!=",
	Between          = "BETWEEN",
	Like             = "LIKE",
	IN               = "IN",
}

const conditionOpArray = ["=",">","<",">=","<=","!=","BETWEEN","LIKE","IN"]


export function GetConditionOps(): string[] {
	return Object.keys(ConditionOp);
}

const numConditionOps = GetConditionOps().length;

export function GetIdxFromConditionOp(op: ConditionOp): number {
	return conditionOpArray.indexOf(op as string);
}

export function StringToConditionOp(s: string): ConditionOp {
	const keyIdx = Object.keys(ConditionOp).indexOf(s)
	if (Object.keys(ConditionOp).indexOf(s) == -1) {
		// TODO: Not in enum, should throw error
	}
	return Object.values(ConditionOp)[keyIdx];
}