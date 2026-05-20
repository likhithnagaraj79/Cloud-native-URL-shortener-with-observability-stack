package com.urlshortener.util;

import org.junit.jupiter.api.Test;
import java.util.HashSet;
import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class ShortCodeGeneratorTest {

    private final ShortCodeGenerator generator = new ShortCodeGenerator();

    @Test
    void generatesCorrectLength() {
        for (int len : new int[]{5, 7, 10, 20}) {
            assertThat(generator.generate(len)).hasSize(len);
        }
    }

    @Test
    void generatesAlphanumericCharactersOnly() {
        for (int i = 0; i < 500; i++) {
            assertThat(generator.generate(7)).matches("[a-zA-Z0-9]+");
        }
    }

    @Test
    void generatesUniqueCodesUnderLoad() {
        Set<String> seen = new HashSet<>(1000);
        for (int i = 0; i < 1000; i++) {
            seen.add(generator.generate(7));
        }
        assertThat(seen).hasSizeGreaterThan(990); // extremely unlikely to have >10 collisions
    }
}
