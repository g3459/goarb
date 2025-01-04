contract CRouter {
    bool internal immutable FRP;
    bool internal immutable GPE;

    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant ADDRESS_MASK = 0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff;
    uint256 internal constant DIREC_MASK = 0x8000000000000000000000000000000000000000000000000000000000000000;
    uint256 internal constant PID_MASK = 0xff000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2_PID = 0x01000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3_PID = 0;
    uint256 internal constant ALGB_PID = 0x02000000000000000000000000000000000000000000000000000000;
    uint256 internal constant VELOV2_PID = 0x03000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2AL_PID = 0x07000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2PK_PID = 0x08000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3PK_PID = 0x05000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3AL_PID = 0x06000000000000000000000000000000000000000000000000000000;
    uint256 internal constant VELOV3_PID = 0x04000000000000000000000000000000000000000000000000000000;
    int24 internal constant MIN_TICK = -887272;
    int24 internal constant MAX_TICK = 887272;

    constructor(bool _FRP, bool _GPE) {
        (FRP, GPE) = (_FRP, _GPE);
    }

    function findRoutes(
        uint8 maxLen,
        uint8 t,
        uint256 amIn,
        bytes[][] memory pools
    )
        public
        view
        returns (
            uint256[] memory amounts,
            bytes[] memory calls,
            uint64[] memory gasUsage
        )
    {
        unchecked {
            amounts = new uint256[](pools.length);
            amounts[t] = amIn;
            calls = new bytes[](pools.length);
            if (GPE) gasUsage = new uint64[](pools.length);
            findRoutes(maxLen * 0x20, pools, amounts, calls, gasUsage);
        }
    }

    function findRoutes(
        uint8 maxLen,
        bytes[][] memory pools,
        uint256[] memory amounts,
        bytes[] memory calls,
        uint64[] memory gasUsage
    ) internal view {
        unchecked {
            uint256 updated = type(uint256).max >> (256 - pools.length);
            while (updated != 0) {
                for (uint256 t0; t0 < pools.length; t0++) {
                    if (updated & (1 << t0) == 0) continue;
                    updated ^= 1 << t0;
                    if (amounts[t0] == 0 || calls[t0].length == maxLen) continue;
                    for (uint256 t1; t1 < pools.length; t1++) {
                        if (t0 == t1) continue;
                        bytes memory _pools;
                        bool direc;
                        if (pools[t0].length > 0 && pools[t0][t1].length > 0) {
                            direc = true;
                            _pools = pools[t0][t1];
                        } else if (pools[t1].length > 0 && pools[t1][t0].length > 0) {
                            _pools = pools[t1][t0];
                        } else {
                            continue;
                        }
                        uint256 eth = t1 == 0 ? 0 : amounts[0];
                        (uint256 hAmOut, uint256 poolCall) = quotePools(amounts[t0] - 2, eth, direc, _pools, calls[t0]);
                        if (hAmOut <= amounts[t1]) continue;
                        if (GPE) {
                            uint256 gasNew = gasUsage[t0] + protGas(poolCall & PID_MASK);
                            {
                                uint256 gasFeeNew = gasNew * tx.gasprice;
                                uint256 gasFeeCurrent = gasUsage[t1] * tx.gasprice;
                                if (eth != 0) {
                                    gasFeeNew = (hAmOut * gasFeeNew) / eth;
                                    gasFeeCurrent = (hAmOut * gasFeeCurrent) / eth;
                                }
                                if (int256(hAmOut - gasFeeNew) <= int256(amounts[t1] - gasFeeCurrent)) continue;
                            }
                            gasUsage[t1] = uint64(gasNew);
                        }
                        amounts[t1] = hAmOut - 2;
                        uint256 amOut56bit = compress56bit(hAmOut - 2);
                        poolCall = (poolCall & (STATE_MASK | PID_MASK | ADDRESS_MASK)) | (amOut56bit << 160);
                        if (direc) poolCall |= DIREC_MASK;
                        calls[t1] = bytes.concat(calls[t0], abi.encode(poolCall));
                        updated |= 1 << t1;
                    }
                }
            }
        }
    }

    // function decompress56bit(uint compressed)internal pure returns (uint){
    //     unchecked{
    //         return uint(uint48(compressed>>8))<<uint8(compressed);
    //     }
    // }

    function quotePools(
        uint256 amIn,
        uint256 eth,
        bool direc,
        bytes memory _pools,
        bytes memory calls
    ) internal view returns (uint256 hAmOut, uint256 poolCall) {
        unchecked {
            uint256 hGasFee;
            for (uint256 p; p < _pools.length; p += 0x40) {
                uint256 slot1;
                assembly {
                    slot1 := mload(add(add(_pools, p), 0x40))
                }
                if (poolInCalls(calls, uint160(slot1))) continue;
                uint256 rIn;
                uint256 rOut;
                {
                    uint256 slot0;
                    assembly {
                        slot0 := mload(add(add(_pools, p), 0x20))
                    }
                    rIn = slot0 >> 128;
                    rOut = uint128(slot0);
                }
                if (!direc) (rIn, rOut) = (rOut, rIn);
                uint256 fee = uint16(slot1 >> 160);
                uint256 amOut = amIn * (1e6 - fee);
                amOut = (amOut * rOut) / (rIn * 1e6 + amOut); ///
                if (amOut <= hAmOut) continue;

                uint256 pid = slot1 & PID_MASK;
                if (pid == UNIV3_PID || pid == ALGB_PID) {
                    int24 s = int24(uint24(uint16(slot1 >> 200)));
                    (int24 tl, int24 tu) = tickBounds(int24(uint24(slot1 >> 176)), s);
                    if (direc ? ((rOut - amOut) << 128) / (rIn + amIn) < tickSqrtPX64(tl)**2 : ((rIn + amIn) << 128) / (rOut - amOut) > tickSqrtPX64(tu)**2) continue;
                }
                if (GPE) {
                    uint256 gasFee = protGas(pid) * tx.gasprice;
                    if (eth != 0) {
                        gasFee = (amOut * gasFee) / eth;
                    }
                    if (int256(amOut - gasFee) <= int256(hAmOut - hGasFee)) continue;
                    if (FRP) {
                        uint256 amOutX2 = (amIn << 1) * (1e6 - fee);
                        amOutX2 = (amOutX2 * rOut) / (rIn * 1e6 + amOutX2);
                        if (int256(amOutX2 - gasFee) > int256((amOut - gasFee) << 1)) continue;
                    }
                    hGasFee = gasFee;
                }
                hAmOut = amOut;
                poolCall = slot1;
            }
        }
    }

    function protGas(uint256 pid) internal pure returns (uint256) {
        unchecked {
            return pid == ALGB_PID ? 300000 : 100000;
        }
    }

    function compress56bit(uint256 uncompressed) internal pure returns (uint256) {
        unchecked {
            uint256 temp = uncompressed;
            uint256 rsh;
            while (uint48(temp) != temp) {
                rsh += 8;
                temp >>= 8;
            }
            temp <<= 8;
            temp |= rsh;
            return temp;
        }
    }

    function poolInCalls(bytes memory calls, uint160 pool) internal pure returns (bool) {
        unchecked {
            for (uint256 i = 0x20; i <= calls.length; i += 0x20) {
                uint256 _poolCall;
                assembly {
                    _poolCall := mload(add(calls, i))
                }
                if (pool == uint160(_poolCall)) {
                    return true;
                }
            }
            return false;
        }
    }

    // function feeAmountTickSpacing(uint fee)internal pure returns(int24 s){
    //     unchecked{
    //         if(fee==100){
    //             return 1;
    //         }
    //         if(fee==500){
    //             return 10;
    //         }
    //         if(fee==2500){
    //             return 50;
    //         }
    //         if(fee==3000){
    //             return 60;
    //         }
    //         if(fee==10000){
    //             return 200;
    //         }
    //     }
    // }

    function tickBounds(int24 t, int24 s) internal pure returns (int24 tl, int24 tu) {
        unchecked {
            assembly {
                tl := mul(sub(sdiv(t, s), and(slt(t, 0), smod(t, s))), s)
            }
            tu = tl + int24(s);
            if (tl < MIN_TICK) tl = MIN_TICK;
            else if (tu > MAX_TICK) tu = MAX_TICK;
        }
    }

    function tickSqrtPX64(int24 tick) internal pure returns (uint256 sqrtPX64) {
        unchecked {
            uint256 absTick;
            assembly {
                tick := signextend(2, tick)
                let mask := sar(255, tick)
                absTick := xor(mask, add(mask, tick))
            }
            uint256 price;
            assembly {
                price := xor(shl(128, 1), mul(xor(shl(128, 1), 0xfffcb933bd6fad37aa2d162d1a594001), and(absTick, 0x1)))
            }
            if (absTick & 0x2 != 0) price = (price * 0xfff97272373d413259a46990580e213a) >> 128;
            if (absTick & 0x4 != 0) price = (price * 0xfff2e50f5f656932ef12357cf3c7fdcc) >> 128;
            if (absTick & 0x8 != 0) price = (price * 0xffe5caca7e10e4e61c3624eaa0941cd0) >> 128;
            if (absTick & 0x10 != 0) price = (price * 0xffcb9843d60f6159c9db58835c926644) >> 128;
            if (absTick & 0x20 != 0) price = (price * 0xff973b41fa98c081472e6896dfb254c0) >> 128;
            if (absTick & 0x40 != 0) price = (price * 0xff2ea16466c96a3843ec78b326b52861) >> 128;
            if (absTick & 0x80 != 0) price = (price * 0xfe5dee046a99a2a811c461f1969c3053) >> 128;
            if (absTick & 0x100 != 0) price = (price * 0xfcbe86c7900a88aedcffc83b479aa3a4) >> 128;
            if (absTick & 0x200 != 0) price = (price * 0xf987a7253ac413176f2b074cf7815e54) >> 128;
            if (absTick & 0x400 != 0) price = (price * 0xf3392b0822b70005940c7a398e4b70f3) >> 128;
            if (absTick & 0x800 != 0) price = (price * 0xe7159475a2c29b7443b29c7fa6e889d9) >> 128;
            if (absTick & 0x1000 != 0) price = (price * 0xd097f3bdfd2022b8845ad8f792aa5825) >> 128;
            if (absTick & 0x2000 != 0) price = (price * 0xa9f746462d870fdf8a65dc1f90e061e5) >> 128;
            if (absTick & 0x4000 != 0) price = (price * 0x70d869a156d2a1b890bb3df62baf32f7) >> 128;
            if (absTick & 0x8000 != 0) price = (price * 0x31be135f97d08fd981231505542fcfa6) >> 128;
            if (absTick & 0x10000 != 0) price = (price * 0x9aa508b5b7a84e1c677de54f3e99bc9) >> 128;
            if (absTick & 0x20000 != 0) price = (price * 0x5d6af8dedb81196699c329225ee604) >> 128;
            if (absTick & 0x40000 != 0) price = (price * 0x2216e584f5fa1ea926041bedfe98) >> 128;
            if (absTick & 0x80000 != 0) price = (price * 0x48a170391f7dc42444e8fa2) >> 128;
            assembly {
                if sgt(tick, 0) {
                    price := div(not(0), price)
                }
                sqrtPX64 := shr(64, price)
            }
        }
    }
}
