type ColorResult = {
    r: number;
    g: number;
    b: number;
    hex: string;
};

type KeyParams = {
    keyR: number;
    keyG: number;
    keyB: number;
    tolerance: number;
    softness: number;
};

/**
 * 支援的原生系統對話框類型。
 */
type DialogType = "info" | "error" | "warning";