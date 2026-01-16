#include "textflag.h"

TEXT ·GetMapType(SB), NOSPLIT, $0-24
    MOVQ m_type+0(FP), AX 
    MOVQ AX, ret+16(FP)   
    RET


