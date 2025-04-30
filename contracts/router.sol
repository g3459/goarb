import "https://github.com/Uniswap/v4-core/blob/main/src/libraries/SqrtPriceMath.sol";
import "https://github.com/Uniswap/v4-core/blob/main/src/libraries/TickMath.sol";

contract CRouter {
    bool internal constant FRP = true;
    bool internal constant GPE = true;
    uint256 internal immutable maxLen;

    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant ADDRESS_MASK = 0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff;
    uint256 internal constant DIREC_MASK = 0x8000000000000000000000000000000000000000000000000000000000000000;
    uint256 internal constant PID_MASK = 0xff;

    uint256 internal constant UNIV3_FID = 1;
    uint256 internal constant VELOV3_FID = 2;
    uint256 internal constant ALGB_FID = 3;
    uint256 internal constant VELOV2_FID = 4;
    uint256 internal constant UNIV2_FID = 5;

    uint256 internal constant FEE_POS = 224;
    uint256 internal constant FID_POS = 240;
    uint256 internal constant PID_POS = 248;

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
    ) public view returns (uint256[] memory amounts, bytes[] memory calls) {
        unchecked {
            amounts = new uint256[](pools.length);
            amounts[tIn] = amIn;
            calls = new bytes[](pools.length);
            findRoutes(pools, amounts, calls);
        }
    }

    function findRoutes(
        bytes[][] calldata pools,
        uint256[] memory amounts,
        bytes[] memory calls
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
                        uint8 callPid = 0xff;
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
                            uint8 pid = uint8(slot >> PID_POS);
                            uint8 fid = uint8(slot >> FID_POS);
                            int24 s = poolSpacing(pid, fid);
                            if (s > 0) {
                                int24 tick = TickMath.getTickAtSqrtPrice(sqrtPriceCurrentX96);
                                int24 tl = tickLower(tick, s);

                                if (direc ? (sqrtPriceNextX96 < TickMath.getSqrtPriceAtTick(tl)) : (sqrtPriceNextX96 >= TickMath.getSqrtPriceAtTick(tl + s))) continue;
                            }
                            uint256 amOut = direc
                                ? SqrtPriceMath.getAmount1Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false)
                                : SqrtPriceMath.getAmount0Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false);
                            if ((!GPE) && amOut <= hamOut) continue;

                            if (GPE) {
                                uint256 _gas = protGas(fid) + gas[t0];
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
                        if (callPid == 0xff) continue;
                        amounts[t1] = hamOut;
                        gas[t1] = hGas;
                        calls[t1] = bytes.concat(bytes2(uint16((t0 << 12) | (t1 << 8) | callPid)), calls[t0]);
                    }
                }
            } while (updated != 0);
        }
    }

    function protGas(uint256 fid) internal pure returns (uint256) {
        unchecked {
            return fid == ALGB_FID ? 300000 : 150000;
        }
    }

    function poolSpacing(uint8 pid, uint8 fid) internal pure returns (int24 s) {
        unchecked {
            if (fid == UNIV3_FID) {
                assembly {
                    switch and(pid, 0x0f)
                    case 0 {
                        s := 1
                    }
                    case 1 {
                        s := 10
                    }
                    case 2 {
                        s := 60
                    }
                    case 3 {
                        s := 200
                    }
                }
            } else if (fid == VELOV3_FID) {
                assembly {
                    switch and(pid, 0x0f)
                    case 0 {
                        s := 1
                    }
                    case 1 {
                        s := 50
                    }
                    case 2 {
                        s := 100
                    }
                    case 3 {
                        s := 200
                    }
                    case 4 {
                        s := 2000
                    }
                }
            } else if (fid == ALGB_FID) {
                s = 60;
            }
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
