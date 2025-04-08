import "https://github.com/Uniswap/v4-core/blob/main/src/libraries/SqrtPriceMath.sol";
import "https://github.com/Uniswap/v4-core/blob/main/src/libraries/TickMath.sol";

contract CRouter {
    bool internal constant FRP = true;
    bool internal constant GPE = true;
    uint256 internal immutable maxLen;

    address internal constant UNIV3_FACTORY = 0x1F98431c8aD98523631AE4a59f267346ea31F984;
    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant ADDRESS_MASK = 0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff;
    uint256 internal constant DIREC_MASK = 0x8000000000000000000000000000000000000000000000000000000000000000;
    uint256 internal constant PID_MASK = 0xff;
    uint256 internal constant UNIV3_PID = 4;
    uint256 internal constant VELOV3_PID = 9;
    uint256 internal constant ALGB_PID = 10;
    uint256 internal constant VELOV2_PID = 12;
    uint256 internal constant UNIV2_PID = 13;
    uint256 internal constant FEE_POS = 176;
    uint256 internal constant SPACING_POS = 200;
    uint256 internal constant PID_POS = 216;
    uint256 internal constant CALLLEN = 16;

    // int24 internal constant MIN_TICK = -887272;
    // int24 internal constant MAX_TICK = 887272;

    constructor(uint256 _maxLen) {
        require(_maxLen != 0);
        maxLen = _maxLen - 1;
    }

    function findRoutes(
        bytes[][] calldata pools,
        uint256 amIn,
        uint8 tIn
    ) public view returns (uint256[] memory amounts, uint256[] memory calls) {
        unchecked {
            amounts = new uint256[](pools.length);
            amounts[tIn] = amIn;
            calls = new uint256[](pools.length);
            findRoutes(pools, amounts, calls);
        }
    }

    function findRoutes(
        bytes[][] calldata pools,
        uint256[] memory amounts,
        uint256[] memory calls
    ) internal view {
        unchecked {
            uint256[] memory gas;
            if (GPE) {
                gas = new uint256[](pools.length);
            }
            uint256 updated = (1 << pools.length) - 1;
            do {
                for (uint256 t0; t0 < pools.length; t0++) {
                    if (updated & (1 << t0) == 0) continue;
                    updated ^= 1 << t0;
                    uint256 amIn = amounts[t0];
                    if (amIn == 0) continue;
                    for (uint256 t1; t1 < pools.length; t1++) {
                        if (t0 == t1) continue;
                        bytes calldata _pools;
                        bool direc;
                        if (pools[t0].length > 0 && pools[t0][t1].length > 0) {
                            direc = true;
                            _pools = pools[t0][t1];
                        } else if (pools[t1].length > 0 && pools[t1][t0].length > 0) {
                            _pools = pools[t1][t0];
                        } else {
                            continue;
                        }
                        uint256 hamOut = amounts[t1];
                        uint256 hGas;
                        if (GPE) {
                            hGas = gas[t1];
                        }
                        uint8 callPid;
                        for (uint256 p; p < _pools.length; p += 0x20) {
                            uint256 slot;
                            assembly {
                                slot := calldataload(add(_pools.offset, p))
                            }
                            uint256 r0 = uint48(slot >> 48);
                            uint256 r1 = uint48(slot);
                            if (r0 == 0 || r1 == 0 || (direc ? r0 : r1) > amIn) continue;
                            uint128 liquidity = uint128(r0 * r1);
                            uint160 sqrtPriceCurrentX96 = uint160((r1 << 96) / r0);
                            uint256 amInLessFee = ((amIn - 2) * (1e6 - uint16(slot >> FEE_POS))) / 1e6;
                            uint160 sqrtPriceNextX96 = SqrtPriceMath.getNextSqrtPriceFromInput(sqrtPriceCurrentX96, liquidity, amInLessFee, direc);
                            int24 s = int24(uint24(uint16(slot >> SPACING_POS)));
                            if (s != 0) {
                                int24 tick = TickMath.getTickAtSqrtPrice(sqrtPriceCurrentX96);
                                int24 tl = tickLower(tick, s);
                                if (direc ? (sqrtPriceNextX96 < TickMath.getSqrtPriceAtTick(tl)) : (sqrtPriceNextX96 >= TickMath.getSqrtPriceAtTick(tl + s))) continue;
                            }
                            uint256 amOut = direc
                                ? SqrtPriceMath.getAmount1Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false)
                                : SqrtPriceMath.getAmount0Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false);
                            uint8 pid = uint8(slot >> PID_POS);
                            if (GPE) {
                                uint256 _gas = protGas(pid) + gas[t0];
                                uint256 gasFee = _gas * tx.gasprice;
                                uint256 hgasFee = hGas * tx.gasprice;
                                if (t1 != 0) {
                                    gasFee = (amOut * gasFee) / amounts[t0];
                                    hgasFee = (amOut * hgasFee) / amounts[t0];
                                }
                                if (amOut - gasFee <= hamOut - hgasFee) continue;
                                if (FRP) {
                                    {
                                        uint160 sqrtPriceNextX96L1 = SqrtPriceMath.getNextSqrtPriceFromInput(sqrtPriceCurrentX96, liquidity, amInLessFee << 1, direc);
                                        uint256 amOutL1 = direc
                                            ? SqrtPriceMath.getAmount1Delta(sqrtPriceCurrentX96, sqrtPriceNextX96L1, liquidity, false)
                                            : SqrtPriceMath.getAmount0Delta(sqrtPriceCurrentX96, sqrtPriceNextX96L1, liquidity, false);
                                        if (int256(amOutL1 - gasFee) > int256((amOut - gasFee) << 1)) continue;
                                    }
                                    {
                                        uint160 sqrtPriceNextX96R1 = SqrtPriceMath.getNextSqrtPriceFromInput(sqrtPriceCurrentX96, liquidity, amInLessFee >> 1, direc);
                                        uint256 amOutR1 = direc
                                            ? SqrtPriceMath.getAmount1Delta(sqrtPriceCurrentX96, sqrtPriceNextX96R1, liquidity, false)
                                            : SqrtPriceMath.getAmount0Delta(sqrtPriceCurrentX96, sqrtPriceNextX96R1, liquidity, false);
                                        if (int256(amOutR1 - gasFee) > int256((amOut - gasFee) >> 1)) continue;
                                    }
                                }
                                hGas = _gas;
                            }
                            callPid = pid;
                            hamOut = amOut;
                        }
                        if (callPid == 0) continue;
                        amounts[t1] = hamOut;
                        gas[t1] = hGas;
                        for (uint256 i; i < 256 / CALLLEN; i += 1) {
                            if (getCall(calls[t0], i) != 0) continue;
                            calls[t0] |= ((t0 << 12) | (t1 << 8) | callPid) << (i * CALLLEN);
                            if(i<=maxLen){
                                updated |= 1 << t1;
                            }
                            break;
                        }
                        
                    }
                }
            } while (updated != 0);
        }
    }

    function getCall(uint256 calls, uint256 ix) internal pure returns (uint256) {
        unchecked {
            return uint16((calls >> (ix * CALLLEN)));
        }
    }

    function protGas(uint256 pid) internal pure returns (uint256) {
        unchecked {
            return (pid & 0x0f) == ALGB_PID ? 300000 : 150000;
        }
    }

    function tickLower(int24 t, int24 s) internal pure returns (int24 tl) {
        unchecked {
            assembly {
                tl := mul(sub(sdiv(t, s), and(slt(t, 0), smod(t, s))), s)
            }
        }
    }
}
